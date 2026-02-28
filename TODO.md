# Done

## Add Tags ✓
Tag recipes (e.g. "braise", "beef", "quick"). Filter the list by tag. When adding a recipe, previously used tags are suggested via a datalist but you can type new tags (comma-separated). Tags are shown on the recipe list and on the recipe page; tag names are clickable to filter by that tag.

## Search By Ingredient ✓
Search box on the list page filters recipes whose ingredient lines contain the search term (e.g. "beef", "flour"). An `ingredients` table stores each used ingredient line so that on the new-recipe form we show "Previously used: ..." to suggest them; you can still type any new ingredient line.

---

# Client-side auto-complete for tags and ingredients ✓
Implemented with Datastar (no JSON): typing in the tags or ingredient-search field triggers a debounced `@get` to `/suggestions/tags` or `/suggestions/ingredients`; the server returns HTML fragments that are patched into `#tag-suggestions` / `#ingredient-suggestions`. Clicking a suggestion updates the bound signal (appends to tags or ingredients). Uses `data-bind`, `data-on:input__debounce.200ms`, and response headers `datastar-selector` / `datastar-mode`.

# Delete a recipe ✓
Delete button on the recipe show page; confirms with `confirm()` then sends `@delete('/recipes/{id}')`. Server deletes and returns SSE redirect to `/recipes`.

# Edit an existing recipe ✓
Edit link on the recipe show page; GET `/recipes/{id}/edit` shows a form pre-filled with the recipe; POST `/recipes/{id}` updates and redirects to the recipe. Edit form has the same tag/ingredient autocomplete as the new-recipe form.

---

(New TODOs can go below.)
