package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Kasjank/skitgubbe/internal/game"
)

func (cfg *apiConfig) handlerCreateGame(w http.ResponseWriter, r *http.Request) {
	user, err := cfg.currentUser(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized", err)
		return
	}

	players := []game.PlayerState{
		{ID: game.PlayerID(user.ID)},
	}

	game := game.NewGame(players)
	cfg.games[game.ID] = game

	fmt.Println(cfg.games)

	respondWithJSON(w, http.StatusOK, struct {
		ID string `json:"id"`
	}{
		ID: game.ID,
	})
}

func (cfg *apiConfig) handlerGamePage(w http.ResponseWriter, r *http.Request) {
	user, err := cfg.currentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	gameID := strings.TrimPrefix(r.URL.Path, "/game/")	

	game, ok := cfg.games[gameID]
	if !ok {
		respondWithError(w, http.StatusNotFound, "Could not find game", err)
		return
	}

	authorized := false	
	for _, player := range game.Players {
		if string(player.ID) == user.ID {
			authorized = true
			break
		}
	}
	if !authorized {
		respondWithError(w, http.StatusForbidden, "Forbidden", err)
		return
	}

	if err := cfg.templates.ExecuteTemplate(w, "game.html", nil); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not execute game template", err)
		return
	}
}

func (cfg *apiConfig) handlerCreateRoom(w http.ResponseWriter, r *http.Request) {
	user, err := cfg.currentUser(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized", err)
		return
	}
	
	room := game.NewRoom(game.PlayerID(user.ID))

	cfg.rooms[room.ID] = room	

	fmt.Println(cfg.rooms)

	respondWithJSON(w, http.StatusOK, struct {
		ID string `json:"id"`
	}{
		ID: room.ID,
	})
}

func (cfg *apiConfig) handlerRoomPage(w http.ResponseWriter, r *http.Request) {
	user, err := cfg.currentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	roomID := strings.TrimPrefix(r.URL.Path, "/room/")	

	room, ok := cfg.rooms[roomID]
	if !ok {
		respondWithError(w, http.StatusNotFound, "Could not find room", err)
		return
	}

	authorized := false	
	for _, player := range room.Players {
		if string(player.ID) == user.ID {
			authorized = true
			break
		}
	}
	if !authorized {
		respondWithError(w, http.StatusForbidden, "Forbidden", err)
		return
	}

	if err := cfg.templates.ExecuteTemplate(w, "room.html", nil); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not execute room template", err)
		return
	}
}

func (cfg *apiConfig) handlerRoomState(w http.ResponseWriter, r *http.Request) {
	user, err := cfg.currentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	roomID := strings.TrimPrefix(r.URL.Path, "/api/rooms/")	

	room, ok := cfg.rooms[roomID]
	if !ok {
		respondWithError(w, http.StatusNotFound, "Could not find room", err)
		return
	}

	authorized := false	
	for _, player := range room.Players {
		if string(player.ID) == user.ID {
			authorized = true
			break
		}
	}
	if !authorized {
		respondWithError(w, http.StatusForbidden, "Forbidden", err)
		return
	}
	fmt.Println(room.Players)	
	respondWithJSON(w, http.StatusOK, room)
}
