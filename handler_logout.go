package main

import "net/http"

func (cfg *apiConfig) handlerLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	delete(cfg.sessions, cookie.Value)

	http.SetCookie(w, &http.Cookie{
		Name: 	"session_id",
		Value:  "",
		Path:   "/",
		HttpOnly: true,
		MaxAge: -1,
	})

	w.WriteHeader(http.StatusOK)
}
