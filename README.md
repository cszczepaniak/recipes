# Recipe App

A simple app to manage personal recipes (ingredients + steps). Built with Go, Templ, Datastar, Tailwind, and SQLite.

## Features

- **List** all recipes on the home page
- **View** a single recipe (ingredients and steps)
- **Create** a new recipe (form with title, ingredients, steps)

## Tech Stack

- **Go** backend with `net/http`
- **SQLite** via [ncruces/go-sqlite3](https://github.com/ncruces/go-sqlite3) (driver + embed)
- **Templ** for HTML templates
- **Datastar** for form submission (no full-page reload)
- **Tailwind** (CDN) for styling

## Run locally

From the repo root:

```bash
# Generate Templ components (after editing .templ files)
templ generate

# Run the server (creates recipes.db in current directory)
go run ./cmd/recipes
```

Then open http://localhost:8080.

### Environment

- `PORT` – HTTP port (default: 8080)
- `DB_PATH` – SQLite file path (default: `recipes.db`)

## Deploy to Railway

The app is a single binary and uses a file-based SQLite database. For Railway, set `DB_PATH` to a path that persists (e.g. under `/data` if using a volume), or use Railway’s persistent disk.

Build:

```bash
go build -o recipes ./cmd/recipes
```

## Tests

```bash
go test ./...
```
