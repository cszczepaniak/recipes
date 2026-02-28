package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateAndGet(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()

	id, err := db.Create(ctx, "Test Recipe", []string{"1 cup flour", "2 eggs"}, []string{"Mix", "Bake at 350°F"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if id != 1 {
		t.Errorf("expected id 1, got %d", id)
	}

	recipe, err := db.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if recipe == nil {
		t.Fatal("expected recipe, got nil")
	}
	if recipe.Title != "Test Recipe" {
		t.Errorf("title: got %q", recipe.Title)
	}
	if len(recipe.Ingredients) != 2 || recipe.Ingredients[0] != "1 cup flour" {
		t.Errorf("ingredients: %v", recipe.Ingredients)
	}
	if len(recipe.Steps) != 2 || recipe.Steps[1] != "Bake at 350°F" {
		t.Errorf("steps: %v", recipe.Steps)
	}
}

func TestList(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()

	_, _ = db.Create(ctx, "First", nil, nil, nil)
	_, _ = db.Create(ctx, "Second", nil, nil, nil)

	list, err := db.List(ctx, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 recipes, got %d", len(list))
	}
	titles := map[string]bool{}
	for _, r := range list {
		titles[r.Title] = true
	}
	if !titles["First"] || !titles["Second"] {
		t.Errorf("expected both recipes in list: %v", list)
	}
}

func TestGetNotFound(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	recipe, err := db.Get(context.Background(), 999)
	if err != nil {
		t.Fatal(err)
	}
	if recipe != nil {
		t.Errorf("expected nil for missing recipe, got %v", recipe)
	}
}

func TestOpenCreatesDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "new.db")
	_, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("database file was not created")
	}
}

func TestCreateWithTags(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()

	id, err := db.Create(ctx, "Tagged Recipe", []string{"flour"}, []string{"Mix"}, []string{"quick", "beef"})
	if err != nil {
		t.Fatal(err)
	}
	tags, err := db.GetRecipeTags(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d: %v", len(tags), tags)
	}
	list, err := db.List(ctx, "quick", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Title != "Tagged Recipe" {
		t.Errorf("list by tag: got %v", list)
	}
}

func TestListByIngredient(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()

	_, _ = db.Create(ctx, "With flour", []string{"2 cups flour", "water"}, []string{"Mix"}, nil)
	_, _ = db.Create(ctx, "No flour", []string{"sugar", "eggs"}, []string{"Mix"}, nil)

	list, err := db.List(ctx, "", "flour")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Title != "With flour" {
		t.Errorf("list by ingredient: got %v", list)
	}
}
