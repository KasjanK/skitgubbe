package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/Kasjank/skitgubbe/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

type apiConfig struct {
	db 	      *database.Queries
	sessions  map[string]string // sessionID -> userID
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
	}

	/* TODO: 
		Create a session + cookie and redirect to your “logged in” page.
	*/

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(filepathRoot)))
	
	mux.HandleFunc("POST /api/login", cfg.handlerLogin)
	mux.HandleFunc("POST /api/signup", cfg.handlerSignup)

	mux.HandleFunc("GET /dashboard", cfg.handlerDashboard)
	mux.HandleFunc("GET /login", cfg.handlerLoginPage)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
	
	defer db.Close()
}

