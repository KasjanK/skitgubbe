package main

import (
	"net/http"
	"strings"

	"github.com/Kasjank/skitgubbe/internal/game"
)

func (cfg *apiConfig) handlerJoinRoom(w http.ResponseWriter, r *http.Request) {
	user, err := cfg.currentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// URL like /api/rooms/{id}/join
    path := strings.TrimPrefix(r.URL.Path, "/api/rooms/")
    parts := strings.SplitN(path, "/", 2)
    if len(parts) != 2 || parts[1] != "join" {
		respondWithError(w, http.StatusNotFound, "room not found", nil)
        return
    }
    roomID := parts[0]

	room, ok := cfg.rooms[roomID]
	if !ok {
		respondWithError(w, http.StatusNotFound, "Could not find room", nil)
		return
	}

	for _, player := range room.Players {
        if string(player.ID) == user.ID {
            respondWithJSON(w, http.StatusOK, struct{ ID string `json:"id"` }{ID: room.ID})
            return
        }
    }

	room.Players = append(room.Players, game.PlayerState{ID: game.PlayerID(user.ID)})

	respondWithJSON(w, http.StatusOK, struct{ ID string `json:"id"` }{ID: room.ID})
}
