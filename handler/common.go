package handler

import (
	"fmt"
	"lightcloud/db"
	"net/http"
	"time"
)

func MainPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/cloud.html")
}

func getSessionUser(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return "", err
	}
	var userID, expiresAtStr string
	err = db.DB.QueryRow("SELECT UserID, ExpiresAt FROM sessions WHERE Token = ?", cookie.Value).Scan(&userID, &expiresAtStr)
	if err != nil {
		return "", err
	}
	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil || time.Now().After(expiresAt) {
		return "", fmt.Errorf("session expired")
	}
	return userID, nil
}
