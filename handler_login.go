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
		Email 	 string `json:"email"`
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

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password", err)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(params.Password))
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password", err)
		return
	}

	//	 TODO:
	//		Generate a random sessionID.
	//		Store sessionID -> userID in a map or DB table.
	//		Set a cookie:
	sessionID := uuid.New()
	cfg.sessions[sessionID.String()] = user.ID

	http.SetCookie(w, &http.Cookie{
		Name: 	"session_id",
		Value:  sessionID.String(),
		Path:   "/",
		HttpOnly: true,
	})

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID: 	user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email: 	   user.Email,
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
