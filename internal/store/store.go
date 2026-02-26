package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

/*
Struct to help store the DB Store properties.
*/
type Store struct {
	db     *sql.DB
	closed bool
	dbPath string
	mutex  sync.Mutex
}

/*
Returns a Store structure that can be used
*/
func Open(dbPath string) (*Store, error) {
	s := &Store{}
	s.mutex.Lock()
	defer s.mutex.Unlock()

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

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
	s.mutex.Lock()
	defer s.mutex.Unlock()

	err := s.db.Close()
	if err != nil {
		return err
	}

	return err
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

	var fixtures []ProblemJson

	var data []byte
	if fixturePath != "" {
		dataT, err := os.ReadFile(fixturePath)
		if err != nil {
			return err
		}
		data = dataT
	} else {
		dataT, err := os.ReadFile("fixtures/catalog.json")
		if err != nil {
			return err
		}
		data = dataT
	}

	err := json.Unmarshal(data, &fixtures)
	if err != nil {
		return err
	}

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

	fmt.Println(query)

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
