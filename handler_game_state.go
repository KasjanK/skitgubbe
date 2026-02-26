package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

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

func (cfg *apiConfig) handlerGameMove(w http.ResponseWriter, r *http.Request) {
	user, err := cfg.currentUser(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized", err)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/games/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 || parts[1] != "move" {
		respondWithError(w, http.StatusNotFound, "bad game path", nil)
		return
	}
	gameID := parts[0]

	gameState, ok := cfg.games[gameID]
	if !ok {
		respondWithError(w, http.StatusNotFound, "game not found", nil)
		return
	}

	var move game.Move
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&move)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not decode move parameters", err)
		return
	}
	
	err = game.ApplyMove(gameState, game.PlayerID(user.ID), move) 	
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid move", err)
		return
	}

	view := game.VisibleStateFor(gameState, game.PlayerID(user.ID))

	if gameState.Finished {
		time.AfterFunc(15 * time.Second, func() { delete(cfg.games, gameState.ID) })
		respondWithError(w, http.StatusNotFound, "Could not find game", nil)
		return
	}

	respondWithJSON(w, http.StatusOK, view)
}
