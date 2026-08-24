package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Store 是所有持久化操作的入口，持有唯一 SQLite 连接。
type Store struct {
	db *sql.DB
}

// Open 打开（或创建）SQLite 数据库并执行迁移。
func Open(path string) (*Store, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// modernc.org/sqlite 单写连接，限制连接池避免 SQLITE_BUSY。
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

// Close 关闭底层连接。
func (s *Store) Close() error {
	return s.db.Close()
}

// DB 暴露底层连接供事务与自检使用。
func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS batches (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			strain TEXT NOT NULL DEFAULT '',
			bioreactor TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS channels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			batch_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			kind TEXT NOT NULL,
			unit TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at TEXT NOT NULL,
			UNIQUE(batch_id, name)
		)`,
		`CREATE TABLE IF NOT EXISTS stages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			batch_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			start_unix INTEGER NOT NULL,
			end_unix INTEGER NOT NULL DEFAULT 0,
			seq INTEGER NOT NULL,
			created_at TEXT NOT NULL,
			UNIQUE(batch_id, seq)
		)`,
		`CREATE TABLE IF NOT EXISTS samples (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			batch_id INTEGER NOT NULL,
			channel_id INTEGER NOT NULL,
			seq INTEGER NOT NULL,
			t_unix INTEGER NOT NULL,
			value REAL NOT NULL,
			created_at TEXT NOT NULL,

		)`,
		`CREATE TABLE IF NOT EXISTS segments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			batch_id INTEGER NOT NULL,
			channel_id INTEGER NOT NULL,
			start_unix INTEGER NOT NULL,
			end_unix INTEGER NOT NULL DEFAULT 0,
			status TEXT NOT NULL,
			lag_secs REAL NOT NULL DEFAULT 0,
			gain REAL NOT NULL DEFAULT 1,
			offset REAL NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS endpoints (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			batch_id INTEGER NOT NULL,
			version INTEGER NOT NULL,
			source TEXT NOT NULL,
			t_unix INTEGER NOT NULL,
			value REAL NOT NULL,
			status TEXT NOT NULL,
			reason TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			UNIQUE(batch_id, version, source)
		)`,
		`CREATE TABLE IF NOT EXISTS snapshots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			batch_id INTEGER NOT NULL,
			version INTEGER NOT NULL,
			status TEXT NOT NULL,
			evidence TEXT NOT NULL DEFAULT '',
			endpoint_t_unix INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			UNIQUE(batch_id, version)
		)`,
	}
	for _, st := range stmts {
		if _, err := s.db.Exec(st); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}
