package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/Kasjank/skitgubbe/internal/database"
	"github.com/Kasjank/skitgubbe/internal/game"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

type apiConfig struct {
	db 	      *database.Queries
	sessions  map[string]string // sessionID -> userID
	games     map[string]*game.GameState
	templates template.Template
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	godotenv.Load(".env")

	pathToDB := os.Getenv("DB_PATH")
	if pathToDB == "" {
		log.Fatal("DB_PATH can't be empty")
	}

	db, err := sql.Open("sqlite3", pathToDB)
	if err != nil {
		log.Fatalf("Could not connect to db: %v", err)
	}

	dbQueries := database.New(db)

	template := template.Must(template.ParseGlob("templates/*.html"))

	cfg := &apiConfig{
		db:	       dbQueries,
		sessions:  make(map[string]string),
		templates: *template,
		games: 	   make(map[string]*game.GameState),
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(filepathRoot)))

	mux.Handle("/static/",
		http.StripPrefix("/static/",
		http.FileServer(http.Dir("static")),
		),
	)
	
	//TODO:
	//POST /api/games/{id}/join  → add logged‑in user to game.
	//GET  /api/games/{id}/state → return view of game state for the current user.
	//POST /api/games/{id}/move  → apply a move.

	mux.HandleFunc("POST /api/login", cfg.handlerLogin)
	mux.HandleFunc("POST /api/signup", cfg.handlerSignup)
	mux.HandleFunc("POST /api/games", cfg.handlerCreateGame)
	

	mux.HandleFunc("GET /dashboard", cfg.handlerDashboard)
	mux.HandleFunc("GET /login", cfg.handlerLoginPage)
	mux.HandleFunc("GET /signup", cfg.handlerSignupPage)
	mux.HandleFunc("GET /game/{id}", cfg.handlerGamePage)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())

	defer db.Close()
}

