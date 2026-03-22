package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Kasjank/skitgubbe/internal/database"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func (cfg apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Username string	`json:"username"` 
		Password string `json:"password"`
	}

	type response struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not decode params", err)
		return 
	}

	user, err := cfg.db.GetUserByUsername(r.Context(), params.Username)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid username or password", err)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(params.Password))
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid username or password", err)
		return
	}

	sessionID := uuid.New()
	cfg.mu.Lock()
	cfg.sessions[sessionID.String()] = user.ID
	cfg.mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name: 	"session_id",
		Value:  sessionID.String(),
		Path:   "/",
		HttpOnly: true,
		Secure: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge: 86400 * 7,
	})

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			Username:  user.Username,
			ID: 	   user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	})
}

func (cfg *apiConfig) currentUser(r *http.Request) (*database.User, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return nil, err
	}

	userID, ok := cfg.sessions[cookie.Value]
	if !ok {
		return nil, fmt.Errorf("invalid session: %v", err)
	}

	user, err := cfg.db.GetUserByID(r.Context(), userID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (cfg *apiConfig) handlerLoginPage(w http.ResponseWriter, r *http.Request) {
	if user, err := cfg.currentUser(r); err == nil && user != nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	if err := cfg.templates.ExecuteTemplate(w, "login.html", nil); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not execute login template", err)
		return
	}
}
