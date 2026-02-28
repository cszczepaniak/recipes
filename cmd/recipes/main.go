package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"recipes/internal/store"
	"recipes/internal/templates"

	"github.com/starfederation/datastar-go/datastar"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "recipes.db"
	}
	db, err := store.Open(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mux := http.NewServeMux()

		// Pages
	mux.HandleFunc("GET /", listHandler(db))
	mux.HandleFunc("GET /recipes", listHandler(db))
	mux.HandleFunc("GET /recipes/new", newRecipeFormHandler(db))
	mux.HandleFunc("GET /api/tags", tagsAPIHandler(db))
	mux.HandleFunc("GET /api/ingredients", ingredientsAPIHandler(db))
	mux.HandleFunc("GET /recipes/{id}", showRecipeHandler(db))
	mux.HandleFunc("POST /recipes", createRecipeHandler(db))

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	log.Printf("Listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func listHandler(db *store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tagFilter := strings.TrimSpace(r.URL.Query().Get("tag"))
		ingredientSearch := strings.TrimSpace(r.URL.Query().Get("ingredient"))
		recipes, err := db.List(r.Context(), tagFilter, ingredientSearch)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tags, _ := db.ListTags(r.Context())
		templates.ListPage(recipes, tags, tagFilter, ingredientSearch).Render(r.Context(), w)
	}
}

func showRecipeHandler(db *store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid recipe id", http.StatusBadRequest)
			return
		}
		recipe, err := db.Get(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if recipe == nil {
			http.NotFound(w, r)
			return
		}
		templates.ShowPage(recipe).Render(r.Context(), w)
	}
}

func newRecipeFormHandler(db *store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tags, _ := db.ListTags(r.Context())
		ingredients, _ := db.ListIngredientNames(r.Context())
		templates.NewRecipePage(tags, ingredients).Render(r.Context(), w)
	}
}

func createRecipeHandler(db *store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		title := strings.TrimSpace(r.FormValue("title"))
		if title == "" {
			http.Error(w, "title required", http.StatusBadRequest)
			return
		}
		ingredients := parseLines(r.FormValue("ingredients"))
		steps := parseLines(r.FormValue("steps"))
		tagNames := parseTagList(r.FormValue("tags"))

		id, err := db.Create(r.Context(), title, ingredients, steps, tagNames)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Datastar: Accept text/event-stream means client expects SSE (e.g. form submitted via Datastar)
		if strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
			sse := datastar.NewSSE(w, r)
			sse.Redirect("/recipes/" + strconv.FormatInt(id, 10))
			return
		}
		http.Redirect(w, r, "/recipes/"+strconv.FormatInt(id, 10), http.StatusSeeOther)
	}
}

func parseLines(s string) []string {
	lines := strings.Split(s, "\n")
	var out []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func parseTagList(s string) []string {
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func tagsAPIHandler(db *store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tags, err := db.ListTags(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		_ = enc.Encode(tags)
	}
}

func ingredientsAPIHandler(db *store.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		names, err := db.ListIngredientNames(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		_ = enc.Encode(names)
	}
}
