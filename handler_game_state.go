package main

import (
	"net/http"
	"strings"

	"github.com/Kasjank/skitgubbe/internal/game"
)

func (cfg *apiConfig) handlerGameState(w http.ResponseWriter, r *http.Request) {
	user, err := cfg.currentUser(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized", err)
		return
	}

	gameID := strings.TrimPrefix(r.URL.Path, "/api/games/")	
	gameState, ok := cfg.games[gameID]
	if !ok {
		respondWithError(w, http.StatusNotFound, "game not found", nil)
		return
	}

	view := game.VisibleStateFor(gameState, game.PlayerID(user.ID))
	respondWithJSON(w, http.StatusOK, view)
}
