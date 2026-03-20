package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/Kasjank/skitgubbe/internal/database"
)

type GameHistoryItem struct {
	GameID 	  string	 `json:"game_id"`
	StartedAt time.Time  `json:"started_at"`
	GameMode  string	 `json:"game_mode,omitempty"`
	Placement int		 `json:"placement"`
}

func (cfg *apiConfig) handlerProfile(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		Username    string 	   		  `json:"username"`
		MemberSince time.Time 		  `json:"member_since"`
		TotalGames  int		  		  `json:"total_games"`
		Wins 		int		  	      `json:"wins"`
		Losses 		int 	  		  `json:"losses"`
		Winrate		int				  `json:"winrate"`
		GameHistory []GameHistoryItem `json:"game_history"`
	}

	_, err := cfg.currentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := strings.TrimPrefix(r.URL.Path, "/api/profile/")
	user, err := cfg.db.GetUserByUsername(r.Context(), username)
	if err != nil { 
		respondWithError(w, http.StatusNotFound, "profile not found", err)
		return
	}

	gameHistory, err := cfg.db.GetMatchHistoryForUser(r.Context(), database.GetMatchHistoryForUserParams{
		UserID: user.ID,
		Limit: 10,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get match history for profile", err)
		return
	}

	var gameHistoryItems []GameHistoryItem
	for _, game := range gameHistory {
		item := GameHistoryItem{
			GameID: game.GameID,
			StartedAt: game.StartedAt,
			GameMode: game.GameMode.String,
			Placement: int(game.Placement),
		}

		gameHistoryItems = append(gameHistoryItems, item)
	}

	stats, err := cfg.db.GetUserStats(r.Context(), user.ID)	
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch stats for profile", err)
		return
	}

	winRate := float64(stats.Wins) / float64(stats.TotalGames) * 100

	respondWithJSON(w, http.StatusOK, Response{
		Username: user.Username,
		MemberSince: user.CreatedAt,
		TotalGames: int(stats.TotalGames),
		Wins: int(stats.Wins),
		Losses: int(stats.Losses),
		Winrate: int(winRate),
		GameHistory: gameHistoryItems,
	})
}

func (cfg *apiConfig) handlerProfilePage(w http.ResponseWriter, r *http.Request) {
	_, err := cfg.currentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := cfg.templates.ExecuteTemplate(w, "profile.html", nil); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not execute profile template", err)
		return
	}
}
