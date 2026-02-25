package store

import (
	"database/sql"
	"sync"

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
