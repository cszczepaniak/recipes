package store

import (
	"context"
	"database/sql"
)

// Recipe is a single recipe with ingredients and steps.
type Recipe struct {
	ID          int64
	Title       string
	Ingredients []string
	Steps       []string
}

// List returns all recipes (id and title only).
func (db *DB) List(ctx context.Context) ([]Recipe, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, title FROM recipes ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Recipe
	for rows.Next() {
		var r Recipe
		if err := rows.Scan(&r.ID, &r.Title); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// Get returns one recipe by ID with ingredients and steps.
func (db *DB) Get(ctx context.Context, id int64) (*Recipe, error) {
	var r Recipe
	err := db.QueryRowContext(ctx, `SELECT id, title FROM recipes WHERE id = ?`, id).Scan(&r.ID, &r.Title)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	// Ingredients
	rows, err := db.QueryContext(ctx, `SELECT line FROM recipe_ingredients WHERE recipe_id = ? ORDER BY ord`, id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			rows.Close()
			return nil, err
		}
		r.Ingredients = append(r.Ingredients, line)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Steps
	rows, err = db.QueryContext(ctx, `SELECT content FROM recipe_steps WHERE recipe_id = ? ORDER BY ord`, id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var content string
		if err := rows.Scan(&content); err != nil {
			rows.Close()
			return nil, err
		}
		r.Steps = append(r.Steps, content)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &r, nil
}

// Create inserts a new recipe with ingredients and steps; returns the new ID.
func (db *DB) Create(ctx context.Context, title string, ingredients, steps []string) (int64, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	res, err := tx.ExecContext(ctx, `INSERT INTO recipes (title) VALUES (?)`, title)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	for i, line := range ingredients {
		if _, err := tx.ExecContext(ctx, `INSERT INTO recipe_ingredients (recipe_id, ord, line) VALUES (?, ?, ?)`, id, i, line); err != nil {
			return 0, err
		}
	}
	for i, content := range steps {
		if _, err := tx.ExecContext(ctx, `INSERT INTO recipe_steps (recipe_id, ord, content) VALUES (?, ?, ?)`, id, i, content); err != nil {
			return 0, err
		}
	}
	return id, tx.Commit()
}
