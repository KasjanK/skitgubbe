package main

import (
	"net/http"

	"github.com/Kasjank/skitgubbe/internal/database"
)

func (cfg *apiConfig) handlerDashboard(w http.ResponseWriter, r *http.Request) {
	user, err := cfg.currentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := struct {
		User *database.User
	}{
		User: user,
	}

	if err := cfg.templates.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not execute dashboard template", err)
		return
	}
}
