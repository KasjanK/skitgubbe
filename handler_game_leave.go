package main

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
)

func (cfg *apiConfig) LeaveGame(gameID string, userID string) error {
    cfg.mu.Lock()
    defer cfg.mu.Unlock()

    gs, err := cfg.getGame(gameID)
    if err != nil {
        return err
    }

    found := false
    for i, p := range gs.Players {
        if string(p.ID) == userID {
            gs.Players = slices.Delete(gs.Players, i, i+1)
			fmt.Printf("%v LEFT FROM GAME %v\n", gameID, gs.ID)
			fmt.Printf("PLAYERS IN GAME: %v", len(gs.Players))
            found = true
            break
        }
    }

    if !found {
        return fmt.Errorf("you are not in this game")
    }

    if len(gs.Players) == 0 {
        delete(cfg.games, gameID)
		fmt.Printf("GAME %v DELETED, NO PLAYERS", gs.ID)
		fmt.Printf("LIVE GAMES: %v\n", cfg.games)
    }

    return nil
}

func (cfg *apiConfig) handlerLeaveGame(w http.ResponseWriter, r *http.Request) {
	user, err := cfg.currentUser(r)
    if err != nil {
        http.Redirect(w, r, "/login", http.StatusSeeOther)
        return
    }
	
    path := strings.TrimPrefix(r.URL.Path, "/api/games/")
    parts := strings.SplitN(path, "/", 2)
    if len(parts) != 2 || parts[1] != "leave" {
		respondWithError(w, http.StatusNotFound, "game not found", nil)
        return
    }
    gameID := parts[0]

	err = cfg.LeaveGame(gameID, user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not leave game", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
