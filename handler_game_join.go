package main

import (
	"net/http"
	"strings"

	"github.com/Kasjank/skitgubbe/internal/game"
)

func (cfg *apiConfig) handlerJoinGame(w http.ResponseWriter, r *http.Request) {
	user, err := cfg.currentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// URL like /api/games/{id}/join
    path := strings.TrimPrefix(r.URL.Path, "/api/games/")
    parts := strings.SplitN(path, "/", 2)
    if len(parts) != 2 || parts[1] != "join" {
		respondWithError(w, http.StatusNotFound, "game not found", nil)
        return
    }
    gameID := parts[0]

	g, ok := cfg.games[gameID]
	if !ok {
		respondWithError(w, http.StatusNotFound, "Could not find game", nil)
		return
	}

	for _, p := range g.Players {
        if string(p.ID) == user.ID {
            respondWithJSON(w, http.StatusOK, struct{ ID string `json:"id"` }{ID: g.ID})
            return
        }
    }

	g.Players = append(g.Players, game.PlayerState{ID: game.PlayerID(user.ID)})

	respondWithJSON(w, http.StatusOK, struct{ ID string `json:"id"` }{ID: g.ID})
}
