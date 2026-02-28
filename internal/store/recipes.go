package store

import (
	"context"
	"database/sql"
	"strings"
)

// Recipe is a single recipe with ingredients, steps, and tags.
type Recipe struct {
	ID          int64
	Title       string
	Ingredients []string
	Steps       []string
	Tags        []string
}

// List returns recipes (id, title, tags). Optional tag filter or ingredient search.
func (db *DB) List(ctx context.Context, tagFilter, ingredientSearch string) ([]Recipe, error) {
	query := `SELECT DISTINCT r.id, r.title FROM recipes r`
	args := []any{}
	if tagFilter != "" {
		query += ` INNER JOIN recipe_tags rt ON rt.recipe_id = r.id INNER JOIN tags t ON t.id = rt.tag_id AND t.name = ?`
		args = append(args, tagFilter)
	}
	if ingredientSearch != "" {
		query += ` INNER JOIN recipe_ingredients ri ON ri.recipe_id = r.id AND ri.line LIKE ?`
		args = append(args, "%"+ingredientSearch+"%")
	}
	query += ` ORDER BY r.created_at DESC`
	rows, err := db.QueryContext(ctx, query, args...)
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
		tags, _ := db.GetRecipeTags(ctx, r.ID)
		r.Tags = tags
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
	r.Tags, _ = db.GetRecipeTags(ctx, id)
	return &r, nil
}

// Create inserts a new recipe with ingredients, steps, and tags; returns the new ID.
func (db *DB) Create(ctx context.Context, title string, ingredients, steps []string, tagNames []string) (int64, error) {
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
	for _, name := range tagNames {
		tagID, err := ensureTagTx(ctx, tx, name)
		if err != nil {
			return 0, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO recipe_tags (recipe_id, tag_id) VALUES (?, ?)`, id, tagID); err != nil {
			return 0, err
		}
	}
	// Record ingredient names for future suggestions
	for _, line := range ingredients {
		line = strings.TrimSpace(line)
		if line != "" {
			_, _ = tx.ExecContext(ctx, `INSERT OR IGNORE INTO ingredients (name) VALUES (?)`, line)
		}
	}
	return id, tx.Commit()
}

// ListTags returns all tag names (for suggestions).
func (db *DB) ListTags(ctx context.Context) ([]string, error) {
	rows, err := db.QueryContext(ctx, `SELECT name FROM tags ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

// ensureTagTx ensures a tag exists and returns its id (must be in transaction).
func ensureTagTx(ctx context.Context, tx *sql.Tx, name string) (int64, error) {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return 0, nil
	}
	res, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO tags (name) VALUES (?)`, name)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	if id != 0 {
		return id, nil
	}
	err = tx.QueryRowContext(ctx, `SELECT id FROM tags WHERE name = ?`, name).Scan(&id)
	return id, err
}

// GetRecipeTags returns tag names for a recipe.
func (db *DB) GetRecipeTags(ctx context.Context, recipeID int64) ([]string, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT t.name FROM tags t INNER JOIN recipe_tags rt ON rt.tag_id = t.id WHERE rt.recipe_id = ? ORDER BY t.name`,
		recipeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

// ListIngredientNames returns all ingredient names (for suggestions).
func (db *DB) ListIngredientNames(ctx context.Context) ([]string, error) {
	rows, err := db.QueryContext(ctx, `SELECT name FROM ingredients ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

// Delete removes a recipe and its ingredients, steps, and tag links.
func (db *DB) Delete(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM recipes WHERE id = ?`, id)
	return err
}

// Update replaces a recipe's title, ingredients, steps, and tags.
func (db *DB) Update(ctx context.Context, id int64, title string, ingredients, steps []string, tagNames []string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `UPDATE recipes SET title = ? WHERE id = ?`, title, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM recipe_ingredients WHERE recipe_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM recipe_steps WHERE recipe_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM recipe_tags WHERE recipe_id = ?`, id); err != nil {
		return err
	}
	for i, line := range ingredients {
		if _, err := tx.ExecContext(ctx, `INSERT INTO recipe_ingredients (recipe_id, ord, line) VALUES (?, ?, ?)`, id, i, line); err != nil {
			return err
		}
	}
	for i, content := range steps {
		if _, err := tx.ExecContext(ctx, `INSERT INTO recipe_steps (recipe_id, ord, content) VALUES (?, ?, ?)`, id, i, content); err != nil {
			return err
		}
	}
	for _, name := range tagNames {
		tagID, err := ensureTagTx(ctx, tx, name)
		if err != nil {
			return err
		}
		if tagID != 0 {
			if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO recipe_tags (recipe_id, tag_id) VALUES (?, ?)`, id, tagID); err != nil {
				return err
			}
		}
	}
	for _, line := range ingredients {
		line = strings.TrimSpace(line)
		if line != "" {
			_, _ = tx.ExecContext(ctx, `INSERT OR IGNORE INTO ingredients (name) VALUES (?)`, line)
		}
	}
	return tx.Commit()
}

// ListTagsMatching returns tag names that contain the query (for autocomplete).
func (db *DB) ListTagsMatching(ctx context.Context, q string) ([]string, error) {
	q = strings.TrimSpace(strings.ToLower(q))
	if q == "" {
		return nil, nil
	}
	rows, err := db.QueryContext(ctx, `SELECT name FROM tags WHERE name LIKE ? ORDER BY name LIMIT 10`, "%"+q+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}

// ListIngredientNamesMatching returns ingredient names that contain the query (for autocomplete).
func (db *DB) ListIngredientNamesMatching(ctx context.Context, q string) ([]string, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil, nil
	}
	rows, err := db.QueryContext(ctx, `SELECT name FROM ingredients WHERE name LIKE ? ORDER BY name LIMIT 10`, "%"+q+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	return out, rows.Err()
}
