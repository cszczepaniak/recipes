# Recipe App

An app to manage personal recipes. Will eventually be deployed to [railway](railway.com)

## Features
0. As simple as possible
1. Create a recipe (ingredients + steps)
2. List all recipes, read a particular recipe

## Tech Stack
- Go for the backend
  - No testify for tests, but write tests as needed!
  - Use the standard library's router (net/http)
- Datastar for interactivity and requests to the backend (https://data-star.dev/)
- Templ for HTML templating (https://github.com/a-h/templ)
  - TemplUI for components if needed (https://templui.io/)
- Tailwind for styling (https://tailwindcss.com/)
- SQLite database in a local file
  - Use the ncruces driver: https://github.com/ncruces/go-sqlite3
- Deployed to railway (https://railway.com)

Use as few dependencies as possible, but if it makes sense to pull one in do it.
