package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Kasjank/skitgubbe/internal/game"
)

func (cfg *apiConfig) CreateRoom(userID game.PlayerID, username string) (*game.Room, error) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	if len(cfg.rooms) > 0 {
		for _, rm := range cfg.rooms {
			for _, player := range rm.Players {
				if player.ID == game.PlayerID(userID) {
					return nil, fmt.Errorf("already in room")
				}
			}
		}
	}
	
	room := game.NewRoom(userID, username)

	cfg.rooms[room.ID] = room
	return room, nil
}

func (cfg *apiConfig) handlerCreateRoom(w http.ResponseWriter, r *http.Request) {
	user, err := cfg.currentUser(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized", err)
		return
	}

	room, err := cfg.CreateRoom(game.PlayerID(user.ID), user.Username)
	if err != nil {
		if err.Error() == "already in room" {
			respondWithError(w, http.StatusConflict, "You are already in a room", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Internal error", err)
		return
	}

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
