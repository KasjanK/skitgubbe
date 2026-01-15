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
