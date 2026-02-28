package store

import (
	"database/sql"
	"fmt"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// DB wraps a sql.DB for the recipe app.
type DB struct {
	*sql.DB
}

// Open opens a SQLite database at path and runs migrations.
func Open(path string) (*DB, error) {
	db, err := sql.Open("sqlite3", "file:"+path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &DB{db}, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS recipes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		);
		CREATE TABLE IF NOT EXISTS recipe_ingredients (
			recipe_id INTEGER NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
			ord INTEGER NOT NULL,
			line TEXT NOT NULL,
			PRIMARY KEY (recipe_id, ord)
		);
		CREATE TABLE IF NOT EXISTS recipe_steps (
			recipe_id INTEGER NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
			ord INTEGER NOT NULL,
			content TEXT NOT NULL,
			PRIMARY KEY (recipe_id, ord)
		);
		CREATE TABLE IF NOT EXISTS tags (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		);
		CREATE TABLE IF NOT EXISTS recipe_tags (
			recipe_id INTEGER NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
			tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
			PRIMARY KEY (recipe_id, tag_id)
		);
		CREATE TABLE IF NOT EXISTS ingredients (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		);
	`)
	if err != nil {
		return err
	}
	// Populate ingredients from existing recipe_ingredients for search suggestions
	_, _ = db.Exec(`
		INSERT OR IGNORE INTO ingredients (name)
		SELECT DISTINCT TRIM(line) FROM recipe_ingredients WHERE TRIM(line) != ''
	`)
	return nil
}
