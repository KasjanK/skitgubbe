package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Kasjank/skitgubbe/internal/database"
	"github.com/Kasjank/skitgubbe/internal/game"
)

func (cfg *apiConfig) getGame(gameID string) (*game.GameState, error) {
    g, ok := cfg.games[gameID]
    if !ok {
        return nil, fmt.Errorf("game not found")
    }
    return g, nil
}

func (cfg *apiConfig) GetGame(id string) (*game.GameState, error) {
    cfg.mu.RLock() 
    defer cfg.mu.RUnlock()
    return cfg.getGame(id)
}

func (cfg *apiConfig) handlerGameState(w http.ResponseWriter, r *http.Request) {
	user, err := cfg.currentUser(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized", err)
		return
	}

	gameID := strings.TrimPrefix(r.URL.Path, "/api/games/")	

	gameState, err := cfg.GetGame(gameID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "game not found", err)
		return
	}

	view := game.VisibleStateFor(gameState, game.PlayerID(user.ID))
	respondWithJSON(w, http.StatusOK, view)
}

func (cfg *apiConfig) SubmitMove(gameID, userID string, move game.Move) (*game.GameState, error) {
    cfg.mu.Lock()
    defer cfg.mu.Unlock()

    gs, err := cfg.getGame(gameID)
    if err != nil {
        return nil, err
    }

    if err := game.ApplyMove(gs, game.PlayerID(userID), move); err != nil {
        return nil, err
    }

    if gs.Finished {
        time.AfterFunc(15*time.Second, func() {
            cfg.mu.Lock()
            delete(cfg.games, gameID)
            cfg.mu.Unlock()
        })
    }

    return gs, nil
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

	var move game.Move
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&move)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Could not decode move parameters", err)
		return
	}
	
	gs, err := cfg.SubmitMove(gameID, user.ID, move)
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "invalid move", err)
        return
    }

	view := game.VisibleStateFor(gs, game.PlayerID(user.ID))

	if gs.Finished {
		err := cfg.db.CreateGame(r.Context(), database.CreateGameParams{
			ID: gs.ID, 
			StartedAt: time.Now(), 
			GameMode: sql.NullString{
				String: "private",
				Valid:  true,  
			},
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Could not insert game into db", err)
			return
		}

		for placement, playerID := range gs.Winners {
			err = cfg.db.AddGameParticipant(r.Context(), database.AddGameParticipantParams{
				GameID: gs.ID, 
				UserID: string(playerID),
				Placement: int64(placement + 1),
			})
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Could not insert participant", err)
				return
			}
		}
		respondWithJSON(w, http.StatusOK, struct{}{})
		return
	}

	respondWithJSON(w, http.StatusOK, view)
}
