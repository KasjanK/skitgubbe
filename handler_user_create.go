package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Kasjank/skitgubbe/internal/database"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID 		  string 	`json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email 	  string	`json:"email"`
}

func (cfg *apiConfig) handlerSignup(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email 	 string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		User
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not decode params", err)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not hash password", err)
		return
	}

	userID := uuid.New()
	user, err := cfg.db.CreateUser(context.Background(), database.CreateUserParams{
		ID: userID.String(),
		Email: params.Email,
		HashedPassword: string(hashedPassword),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create user", err)
		return 
	}

	respondWithJSON(w, http.StatusCreated, response{
		User: User{
			ID: 	   user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email: params.Email,
		},
	})
}

func (cfg *apiConfig) handlerSignupPage(w http.ResponseWriter, r *http.Request) {
	if err := cfg.templates.ExecuteTemplate(w, "signup.html", nil); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not execute signup template", err)
		return
	}
}
