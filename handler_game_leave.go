package main

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/Kasjank/skitgubbe/internal/game"
)

func (cfg *apiConfig) handlerLeaveGame(w http.ResponseWriter, r *http.Request) {
	user, err := cfg.currentUser(r)
    if err != nil {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }
	
	// URL like /api/games/{id}/leave
    path := strings.TrimPrefix(r.URL.Path, "/api/games/")
    parts := strings.SplitN(path, "/", 2)
    if len(parts) != 2 || parts[1] != "leave" {
		respondWithError(w, http.StatusNotFound, "game not found", nil)
        return
    }
    gameID := parts[0]

	gameState := cfg.games[gameID]

	found := false
    for i, player := range gameState.Players {
        if player.ID == game.PlayerID(user.ID) {
            gameState.Players = slices.Delete(gameState.Players, i, i + 1)
            fmt.Printf("%v LEFT FROM GAME %v\n", user.ID, gameState.ID)
			fmt.Printf("PLAYERS IN GAME: %v", len(gameState.Players))
            found = true
            break
        }
    }

	if !found {
        respondWithError(w, http.StatusBadRequest, "You are not in this game", nil)
        return
    }

	if len(gameState.Players) == 0 {
		delete(cfg.games, gameState.ID)
		fmt.Printf("GAME %v DELETED, NO PLAYERS", gameState.ID)
		fmt.Printf("LIVE GAMES: %v\n", cfg.games)
	}

	w.WriteHeader(http.StatusOK)
}
