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
	sessions  map[string]string  		 // sessionID -> userID
	games     map[string]*game.GameState // gameID    -> gamestate
	templates template.Template
	rooms 	  map[string]*game.Room      // roomID    -> room
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
		rooms: 	   make(map[string]*game.Room),
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(filepathRoot)))

	mux.Handle("/static/",
		http.StripPrefix("/static/",
		http.FileServer(http.Dir("static")),
		),
	)
	
	//TODO:
	// - be able to select cards when playing face up/down
	// - if you can play a card from faceup cards, you must pick that card. disable the other cards.
	
	//BUGS: 

	mux.HandleFunc("POST /api/login", cfg.handlerLogin)
	mux.HandleFunc("POST /api/signup", cfg.handlerSignup)

	mux.HandleFunc("POST /api/rooms", cfg.handlerCreateRoom)
	mux.HandleFunc("POST /api/rooms/", cfg.handlerRoomsPost)

	mux.HandleFunc("POST /api/games/", cfg.handlerGameMove)

	mux.HandleFunc("GET /dashboard", cfg.handlerDashboard)
	mux.HandleFunc("GET /login", cfg.handlerLoginPage)
	mux.HandleFunc("GET /signup", cfg.handlerSignupPage)

	mux.HandleFunc("GET /game/", cfg.handlerGamePage)
	mux.HandleFunc("GET /api/games/", cfg.handlerGameState)

	mux.HandleFunc("GET /room/", cfg.handlerRoomPage)
	mux.HandleFunc("GET /api/rooms/", cfg.handlerRoomState)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())

	defer db.Close()
}

