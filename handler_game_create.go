package main

import (
	"fmt"
	"net/http"

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
	//TODO:
	// Only allow players that are connected to the match
	if err := cfg.templates.ExecuteTemplate(w, "game.html", nil); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not execute game template", err)
		return
	}
}
