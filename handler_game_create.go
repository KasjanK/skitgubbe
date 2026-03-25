package main

import (
	"net/http"
	"strings"

)

func (cfg *apiConfig) handlerGamePage(w http.ResponseWriter, r *http.Request) {
	user, err := cfg.currentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	gameID := strings.TrimPrefix(r.URL.Path, "/game/")	

	game, err := cfg.GetGame(gameID)
    if err != nil {
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
