package store

import (
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"math/rand"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

const RAND_SEED = 42069 // NICE

//go:embed fixtures/catalog.json
var defaultFixtures embed.FS

/*
Struct to help store the DB Store properties.
*/
type Store struct {
	db            *sql.DB
	workspacePath string
	closed        bool
	dbPath        string
}

/*
Returns a Store structure that can be used
*/
func Open(dbPath, workspacePath string) (*Store, error) {
	s := &Store{}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	_, err = os.Stat(workspacePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, errors.New("workspace path is invalid")
	}
	s.workspacePath = workspacePath

	var pramgaStmts = []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA busy_timeout = 5000",
	}
	for _, stmt := range pramgaStmts {
		_, err := db.Exec(stmt)
		if err != nil {
			return nil, err
		}
	}

	s.db = db
	s.closed = false
	s.dbPath = dbPath

	if err := s.initSchema(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) initSchema() error {
	var initSchemaStmts = []string{
		`CREATE TABLE IF NOT EXISTS problems (
		id           TEXT PRIMARY KEY,          -- uuid
		source       TEXT NOT NULL,
		source_id    TEXT NOT NULL,
		title        TEXT NOT NULL,
		url          TEXT NOT NULL,
		difficulty   TEXT NOT NULL,
		rating       INTEGER NULL,
		topics       TEXT NOT NULL,             -- JSON array
		tags         TEXT NOT NULL,             -- JSON array
		statement_md TEXT NULL,
		samples_json TEXT NULL,                 -- JSON: [{in, out, name}]
		created_at   INTEGER NOT NULL,
		updated_at   INTEGER NOT NULL,
		UNIQUE(source, source_id)
	);`,

		`CREATE TABLE IF NOT EXISTS user_prefs (
		id                 TEXT PRIMARY KEY,    -- uuid (single profile v0)
		preferred_language TEXT NOT NULL,
		default_topics     TEXT NOT NULL,       -- JSON array
		difficulty_mode    TEXT NOT NULL,
		daily_target_count INTEGER NOT NULL,
		exclude_days       INTEGER NOT NULL DEFAULT 14,
		created_at         INTEGER NOT NULL,
		updated_at         INTEGER NOT NULL
	);`,

		`CREATE TABLE IF NOT EXISTS daily_sets (
		id              TEXT PRIMARY KEY,       -- uuid
		date            TEXT NOT NULL,          -- YYYY-MM-DD
		topics          TEXT NOT NULL,          -- JSON array
		difficulty_mode TEXT NOT NULL,
		target_count    INTEGER NOT NULL,
		status          TEXT NOT NULL,          -- building/ready/failed
		seed            TEXT NOT NULL,          -- date + user_prefs.id (single-profile v0)
		created_at      INTEGER NOT NULL,
		UNIQUE(date)
	);`,

		`CREATE TABLE IF NOT EXISTS daily_set_items (
		daily_set_id TEXT NOT NULL,
		problem_id   TEXT NOT NULL,
		order_index  INTEGER NOT NULL,
		rationale    TEXT NULL,
		created_at   INTEGER NOT NULL,
		PRIMARY KEY(daily_set_id, problem_id),
		FOREIGN KEY(daily_set_id) REFERENCES daily_sets(id) ON DELETE CASCADE,
		FOREIGN KEY(problem_id)   REFERENCES problems(id)   ON DELETE CASCADE
	);`,

		`CREATE TABLE IF NOT EXISTS attempts (
		id                 TEXT PRIMARY KEY,    -- uuid
		problem_id         TEXT NOT NULL,
		daily_set_id       TEXT NULL,
		started_at         INTEGER NULL,
		finished_at        INTEGER NULL,
		status             TEXT NOT NULL,       -- attempted/solved/skipped
		verdict            TEXT NULL,           -- AC/WA/TLE/etc
		notes              TEXT NULL,
		language           TEXT NOT NULL,
		time_spent_seconds INTEGER NOT NULL DEFAULT 0,
		created_at         INTEGER NOT NULL,
		FOREIGN KEY(problem_id)  REFERENCES problems(id)   ON DELETE CASCADE,
		FOREIGN KEY(daily_set_id) REFERENCES daily_sets(id) ON DELETE SET NULL
	);`,

		`CREATE INDEX IF NOT EXISTS idx_problems_difficulty
		ON problems(difficulty);`,

		`CREATE INDEX IF NOT EXISTS idx_attempts_problem_created
		ON attempts(problem_id, created_at);`,

		`CREATE INDEX IF NOT EXISTS idx_daily_set_items_order
		ON daily_set_items(daily_set_id, order_index);`,
	}

	for _, stmt := range initSchemaStmts {
		_, err := s.db.Exec(stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
Closes the database connection
*/
func (s *Store) Close() error {
	err := s.db.Close()
	if err != nil {
		return err
	}

	return err
}

/*
Fetch the workspace path
*/
func (s *Store) WorkspacePath() string {
	return s.workspacePath
}

/*
Returns the total count of problems in the store
*/
func (s *Store) CountProblems() (int, error) {
	var count int

	row := s.db.QueryRow("SELECT COUNT(*) FROM problems")
	if err := row.Scan(&count); err != nil {
		return -1, err
	}

	return count, nil
}

/*
Upserts fixtures into the db from the fixture/catalog.json or from
the import flag.
*/
func (s *Store) UpsertProblemsFromFixture(fixturePath string) error {
	tx, _ := s.db.Begin()
	source := rand.NewSource(RAND_SEED)
	r := rand.New(source)

	var fixtures []ProblemJson

	var data []byte
	if fixturePath != "" {
		dataT, err := os.ReadFile(fixturePath)
		if err != nil {
			return err
		}
		data = dataT
	} else {
		// dataT, err := os.ReadFile("fixtures/catalog.json")
		dataT, err := fs.ReadFile(defaultFixtures, "fixtures/catalog.json")
		if err != nil {
			return err
		}
		data = dataT
	}

	err := json.Unmarshal(data, &fixtures)
	if err != nil {
		return err
	}

	r.Shuffle(len(fixtures), func(i, j int) {
		fixtures[i], fixtures[j] = fixtures[j], fixtures[i]
	})

	for _, fixture := range fixtures {
		fixtureID := uuid.New().String()
		createdAt := time.Now().Unix()
		updatedAt := createdAt

		topicsBytes, err := json.Marshal(fixture.Topics)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		tagsBytes, err := json.Marshal(fixture.Tags)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
		samplesBytes, err := json.Marshal(fixture.Samples)
		if err != nil {
			_ = tx.Rollback()
			return err
		}

		queryString := `
		INSERT INTO problems(
			id, source, source_id, title, url, difficulty, rating, topics, tags, statement_md, samples_json, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(source, source_id) DO UPDATE SET
				title = excluded.title,
				url = excluded.url,
				difficulty = excluded.difficulty,
				rating = excluded.rating,
				topics = excluded.topics,
				tags = excluded.tags,
				statement_md = excluded.statement_md,
				samples_json = excluded.samples_json,
				updated_at = excluded.updated_at
		`

		_, err = tx.Exec(
			queryString,
			fixtureID,
			fixture.Source,
			fixture.SourceID,
			fixture.Title,
			fixture.URL,
			fixture.Difficulty,
			fixture.Rating,
			string(topicsBytes),
			string(tagsBytes),
			fixture.StatementMd,
			string(samplesBytes),
			createdAt,
			updatedAt,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return err
	}

	return nil
}

/*
This function give a list of problems that is currently stored
in our database
*/
func (s *Store) ListProblems(uf UserFilters) ([]Problem, error) {
	query := `SELECT id, source, title, url, difficulty, rating, topics, tags, created_at FROM problems`
	var whereClauses []string
	var args []interface{}

	if uf.Source != "" && slices.Contains([]string{"codeforces", "leetcode"}, uf.Source) {
		whereClauses = append(whereClauses, "source = ?")
		args = append(args, uf.Source)
	}

	if uf.Difficulty != "" {
		whereClauses = append(whereClauses, "difficulty = ?")
		args = append(args, uf.Difficulty)
	}

	if uf.Topic != "" {
		whereClauses = append(whereClauses, "topics LIKE ?")
		args = append(args, "%\""+uf.Topic+"\"%")
	}

	if uf.Tag != "" {
		whereClauses = append(whereClauses, "tags LIKE ?")
		args = append(args, "%\""+uf.Tag+"\"%")
	}

	if uf.Title != "" {
		whereClauses = append(whereClauses, "title LIKE ?")
		args = append(args, "%"+uf.Title+"%")
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	if uf.Limit > 0 {
		query += " LIMIT ? "
		args = append(args, uf.Limit)
	}

	if uf.Offset > 0 {
		query += " OFFSET ? "
		args = append(args, uf.Offset)
	}

	var fetchedProblems []Problem

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p Problem

		var topicsString string
		var tagsString string

		if err := rows.Scan(
			&p.Id, &p.Source,
			&p.Title, &p.Url, &p.Difficulty,
			&p.Rating, &topicsString, &tagsString,
			&p.CreatedAt,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(topicsString), &p.Topics); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(tagsString), &p.Tags); err != nil {
			return nil, err
		}

		fetchedProblems = append(fetchedProblems, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return fetchedProblems, nil
}

/*
This function returns a Problem struct based on the given
ID - Marshal the required fields too.
*/
func (s *Store) GetProblemByID(id string) (ProblemID, error) {
	var pID ProblemID

	var topicString string
	var tagString string
	var sampleString string

	row := s.db.QueryRow("SELECT id, source, source_id, title, url, difficulty, rating, topics, tags, statement_md, samples_json, created_at, updated_at FROM problems WHERE id = ?", id)
	if row.Err() != nil {
		return ProblemID{}, row.Err()
	}

	if err := row.Scan(
		&pID.Id, &pID.Source, &pID.SourceID, &pID.Title,
		&pID.Url, &pID.Difficulty, &pID.Rating,
		&topicString, &tagString, &pID.StatementMd,
		&sampleString, &pID.CreatedAt, &pID.updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ProblemID{}, errors.New("no problem with id " + id + " found")
		}
		return ProblemID{}, err
	}

	if err := json.Unmarshal([]byte(topicString), &pID.Topics); err != nil {
		return ProblemID{}, err
	}

	if err := json.Unmarshal([]byte(tagString), &pID.Tags); err != nil {
		return ProblemID{}, err
	}

	if err := json.Unmarshal([]byte(sampleString), &pID.Samples); err != nil {
		return ProblemID{}, err
	}

	return pID, nil
}

/*
Helper to count the problems with the set filter
*/
func (s *Store) CountProblemsWithFilters(uf UserFilters) (int, error) {
	query := `SELECT COUNT(*) FROM problems`
	var whereClauses []string
	var args []interface{}

	if uf.Source != "" && slices.Contains([]string{"codeforces", "leetcode"}, uf.Source) {
		whereClauses = append(whereClauses, "source = ?")
		args = append(args, uf.Source)
	}

	if uf.Difficulty != "" {
		whereClauses = append(whereClauses, "difficulty = ?")
		args = append(args, uf.Difficulty)
	}

	if uf.Topic != "" {
		whereClauses = append(whereClauses, "topics LIKE ?")
		args = append(args, "%\""+uf.Topic+"\"%")
	}

	if uf.Tag != "" {
		whereClauses = append(whereClauses, "tags LIKE ?")
		args = append(args, "%\""+uf.Tag+"\"%")
	}

	if uf.Title != "" {
		whereClauses = append(whereClauses, "title LIKE ?")
		args = append(args, "%"+uf.Title+"%")
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	var count int
	row := s.db.QueryRow(query, args...)
	if err := row.Scan(&count); err != nil {
		return -1, err
	}
	return count, nil
}

/*
This function is required to insert an attempt into the
sqlite db when a user hits "r"
*/
func (s *Store) InsertAttempt(in CreateAttemptInput) error {
	if s == nil || s.db == nil {
		return errors.New("store is nil")
	}
	if in.ProblemID == "" {
		return errors.New("problem id is required")
	}

	id := uuid.NewString()
	createdAt := time.Now().Unix()

	_, err := s.db.Exec(`
		INSERT INTO attempts (
			id,
			problem_id,
			daily_set_id,
			started_at,
			finished_at,
			status,
			verdict,
			notes,
			language,
			time_spent_seconds,
			created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		id,
		in.ProblemID,
		in.DailySetID,
		in.StartedAt,
		in.FinishedAt,
		in.Status,
		in.Verdict,
		in.Notes,
		in.Language,
		in.TimeSpentSeconds,
		createdAt,
	)
	return err
}

/*
This function will list all the atetmpts for a given
problem ID
*/
func (s *Store) ListAttemptsByProblemID(problemID string, limit int) ([]Attempt, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("store is nil")
	}
	if problemID == "" {
		return nil, errors.New("problem id is required")
	}
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.db.Query(`
		SELECT
			id,
			problem_id,
			daily_set_id,
			started_at,
			finished_at,
			status,
			verdict,
			notes,
			language,
			time_spent_seconds,
			created_at
		FROM attempts
		WHERE problem_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`, problemID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Attempt
	for rows.Next() {
		var a Attempt
		var dailySetID sql.NullString
		var startedAt sql.NullInt64
		var finishedAt sql.NullInt64

		err := rows.Scan(
			&a.ID,
			&a.ProblemID,
			&dailySetID,
			&startedAt,
			&finishedAt,
			&a.Status,
			&a.Verdict,
			&a.Notes,
			&a.Language,
			&a.TimeSpentSeconds,
			&a.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if dailySetID.Valid {
			v := dailySetID.String
			a.DailySetID = &v
		}
		if startedAt.Valid {
			v := startedAt.Int64
			a.StartedAt = &v
		}
		if finishedAt.Valid {
			v := finishedAt.Int64
			a.FinishedAt = &v
		}

		out = append(out, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

/*
This function will list all the atetmpts
*/
func (s *Store) ListAllAttempts() ([]Attempt, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("store is nil")
	}

	rows, err := s.db.Query(`
		SELECT *
		FROM (
    		SELECT *,
        	ROW_NUMBER() OVER (
				PARTITION BY problem_id ORDER BY created_at DESC
			) as rn
    		FROM 
        	attempts
		) AS subquery
		WHERE 
    	rn = 1;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Attempt
	for rows.Next() {
		var a Attempt
		var dailySetID sql.NullString
		var startedAt sql.NullInt64
		var finishedAt sql.NullInt64
		var rn string
		err := rows.Scan(
			&a.ID,
			&a.ProblemID,
			&dailySetID,
			&startedAt,
			&finishedAt,
			&a.Status,
			&a.Verdict,
			&a.Notes,
			&a.Language,
			&a.TimeSpentSeconds,
			&a.CreatedAt,
			&rn,
		)
		if err != nil {
			return nil, err
		}

		if dailySetID.Valid {
			v := dailySetID.String
			a.DailySetID = &v
		}
		if startedAt.Valid {
			v := startedAt.Int64
			a.StartedAt = &v
		}
		if finishedAt.Valid {
			v := finishedAt.Int64
			a.FinishedAt = &v
		}

		out = append(out, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}
