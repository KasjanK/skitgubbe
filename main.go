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

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(filepathRoot)))
	// TODO:
	// mux.HandleFunc("POST /api/login", cfg.handlerLogin)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
	
	defer db.Close()
}

