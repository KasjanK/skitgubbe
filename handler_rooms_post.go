package main

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

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
		cfg.handleJoinRoomAction(w, user, roomID)
	case "start":
		cfg.handleStartRoomAction(w, r, user, roomID)
	case "leave":
		cfg.handleLeaveRoomAction(w, r, user, roomID)
	default:
		respondWithError(w, http.StatusNotFound, "unknown rooms action", nil)
	}
}

func (cfg *apiConfig) JoinRoom(roomID, userID, username string) error {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	room, ok := cfg.rooms[roomID]
	if !ok { 
		return fmt.Errorf("room not found") 
	}

	for _, rm := range cfg.rooms {
        for _, p := range rm.Players {
            if string(p.ID) == userID { 
				return fmt.Errorf("already in room") 
			}
        }
    }

	room.Players = append(room.Players, game.PlayerState{ Username: username, ID: game.PlayerID(userID)})
	return nil
}

func (cfg *apiConfig) handleJoinRoomAction(w http.ResponseWriter, user *database.User, roomID string) {
	err := cfg.JoinRoom(roomID, user.ID, user.Username)
	if err != nil {
		if err.Error() == "room not found" {
			respondWithError(w, http.StatusNotFound, "Room not found", nil)
			return
		}
		if err.Error() == "already in room" {
			respondWithError(w, http.StatusConflict, "You are already in a room", nil)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Internal error", nil)
        return
	}

	respondWithJSON(w, http.StatusOK, struct{ ID string `json:"id"` }{ID: roomID})
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

	fmt.Println("LIVE GAMES:")
	for id, state := range cfg.games {
		fmt.Println(id, state.ID)
	}

	if room.Started {
		time.AfterFunc(15 * time.Second, func() { delete(cfg.rooms, room.ID) })
	}

	fmt.Println("LIVE ROOMS:")
	for id, room := range cfg.rooms {
		fmt.Println(id, room.ID)
	}
	respondWithJSON(w, http.StatusOK, struct{ GameID string `json:"game_id"` }{GameID: g.ID})
}

func (cfg *apiConfig) handleLeaveRoomAction(w http.ResponseWriter, r *http.Request, user *database.User, roomID string) {
	room, ok := cfg.rooms[roomID]
	if !ok {
		respondWithError(w, http.StatusNotFound, "room not found", nil)
		return
	}

	found := false
    for i, player := range room.Players {
        if player.ID == game.PlayerID(user.ID) {
            room.Players = slices.Delete(room.Players, i, i + 1)
            fmt.Printf("%v LEFT FROM ROOM %v\n", user.ID, roomID)
            found = true
            break
        }
    }

	if !found {
        respondWithError(w, http.StatusBadRequest, "You are not in this room", nil)
        return
    }

	if len(room.Players) == 0 {
        delete(cfg.rooms, roomID)
		fmt.Printf("ROOM %v DELETED, NO PLAYERS\n", roomID)
    }

	respondWithJSON(w, http.StatusOK, struct{ ID string `json:"id"` }{ID: room.ID})
}

