package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (cfg *apiConfig) handlerProfile(w http.ResponseWriter, r *http.Request) {
	type Response struct {
		Username    string 	  `json:"username"`
		MemberSince time.Time `json:"member_since"`
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

	fmt.Println(username, user.Username)

	respondWithJSON(w, http.StatusOK, Response{
		Username: user.Username,
		MemberSince: user.CreatedAt,
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
