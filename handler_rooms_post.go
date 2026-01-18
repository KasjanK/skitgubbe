package main

import (
	"net/http"
	"strings"

	"github.com/Kasjank/skitgubbe/internal/database"
	"github.com/Kasjank/skitgubbe/internal/game"
)

func (cfg *apiConfig) handlerRoomsPost(w http.ResponseWriter, r *http.Request) {
	user, err := cfg.currentUser(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized", err)
		return
	}

	tail := strings.TrimPrefix(r.URL.Path, "/api/rooms/")

	parts := strings.SplitN(tail, "/", 2)
	if len(parts) != 2 {
		respondWithError(w, http.StatusNotFound, "bad rooms path", nil)
		return
	}
	roomID, action := parts[0], parts[1]

	switch action {
	case "join":
		cfg.handleJoinRoomAction(w, r, user, roomID)
	case "start":
		cfg.handleStartRoomAction(w, r, user, roomID)
	default:
		respondWithError(w, http.StatusNotFound, "unknown rooms action", nil)
	}
}

func (cfg *apiConfig) handleJoinRoomAction(w http.ResponseWriter, r *http.Request, user *database.User, roomID string) {
	room, ok := cfg.rooms[roomID]
	if !ok {
		respondWithError(w, http.StatusNotFound, "room not found", nil)
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

func (cfg *apiConfig) handleStartRoomAction(w http.ResponseWriter, r *http.Request, user *database.User, roomID string) {
	room, ok := cfg.rooms[roomID]
	if !ok {
		respondWithError(w, http.StatusNotFound, "room not found", nil)
		return
	}

	if string(room.OwnerID) != user.ID {
		respondWithError(w, http.StatusForbidden, "only host can start the match", nil)
		return
	}

	g := game.NewGame(room.Players)
	cfg.games[g.ID] = g
	room.Started = true
	room.GameID = g.ID

	respondWithJSON(w, http.StatusOK, struct{ GameID string `json:"game_id"` }{GameID: g.ID})
}
