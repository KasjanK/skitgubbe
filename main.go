package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/Kasjank/skitgubbe/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

type apiConfig struct {
	db 	*database.Queries
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

	cfg := apiConfig{
		db:	dbQueries,
	}

	/* TODO: 
		Parse form/JSON for email + password.
		Check if a user with that email already exists (or catch the unique‑constraint error).
		Hash the password with bcrypt.
		Generate a UUID.
		Call CreateUser.
		Create a session + cookie and redirect to your “logged in” page.
	*/

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(filepathRoot)))
	
	mux.HandleFunc("POST /api/login", cfg.handlerLogin)
	mux.HandleFunc("POST /api/signup", cfg.handlerSignup)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
	
	defer db.Close()
}

