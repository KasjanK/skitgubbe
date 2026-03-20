package main

import (
	"database/sql"
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/Kasjank/skitgubbe/internal/database"
	"github.com/Kasjank/skitgubbe/internal/game"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

type apiConfig struct {
	db 	      *database.Queries
	sessions  map[string]string  		 // sessionID -> userID
	games     map[string]*game.GameState // gameID    -> gamestate
	templates template.Template
	rooms 	  map[string]*game.Room      // roomID    -> room
}
//go:embed templates/*.html
var templateFS embed.FS

//go:embed static/* 
var staticFS embed.FS

//go:embed sql/schema/*.sql
var migrationFS embed.FS

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

	template, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		panic(err)
	}

	if err := goose.SetDialect("sqlite3"); err != nil {
   		panic(err)
	}

    subFS, err := fs.Sub(migrationFS, "sql/schema")
    if err != nil {
        panic(err)
    }

	goose.SetBaseFS(subFS)
    if err := goose.Up(db, "."); err != nil {
        panic(err)
    }

	cfg := &apiConfig{
		db:	       dbQueries,
		sessions:  make(map[string]string),
		templates: *template,
		games: 	   make(map[string]*game.GameState),
		rooms: 	   make(map[string]*game.Room),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		err := template.ExecuteTemplate(w, "index.html", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.Handle("/static/", http.FileServer(http.FS(staticFS)))
	
	//TODO:
	// - GET PLACE IN MATCH HISTORY
	// - reassign owner when owner leave
	// - remove ready function
	// - only show start button for owner of room
	// - add player info in room
	// - if player is in game, redirect from dashboard to game

	//BUGS: 

	mux.HandleFunc("POST /api/login", cfg.handlerLogin)
	mux.HandleFunc("POST /api/signup", cfg.handlerSignup)
	mux.HandleFunc("POST /api/logout", cfg.handlerLogout)

	mux.HandleFunc("POST /api/rooms", cfg.handlerCreateRoom)
	mux.HandleFunc("POST /api/rooms/", cfg.handlerRoomsPost)

	mux.HandleFunc("POST /api/games/", cfg.handlerGameMove)
	mux.HandleFunc("DELETE /api/games/", cfg.handlerLeaveGame)

	mux.HandleFunc("GET /dashboard", cfg.handlerDashboard)
	mux.HandleFunc("GET /login", cfg.handlerLoginPage)
	mux.HandleFunc("GET /signup", cfg.handlerSignupPage)

	mux.HandleFunc("GET /profile/", cfg.handlerProfilePage)
	mux.HandleFunc("GET /api/profile/", cfg.handlerProfile)

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

