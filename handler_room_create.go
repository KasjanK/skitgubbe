package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Kasjank/skitgubbe/internal/game"
)

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

type RoomPageData struct {
	RoomID string
	UserID string
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

	data := RoomPageData{
		RoomID: room.ID,
		UserID: user.ID,
	}

	if err := cfg.templates.ExecuteTemplate(w, "room.html", data); err != nil {
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
	respondWithJSON(w, http.StatusOK, room)
}
