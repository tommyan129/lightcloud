package handler

import (
	"crypto/rand"
	"database/sql"
	"log"
	"net/http"
	"time"

	"lightcloud/db"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func generateShareToken() string {
	token := make([]byte, 16)
	_, err := rand.Read(token)
	if err != nil {
		log.Fatalf("ID 생성 실패: %v\n", err)
	}
	resTok := make([]byte, 0, 16)
	for _, t := range token {

		resTok = append(resTok, base62Chars[t%62])
	}
	return string(resTok)
}

func CreateShareLink(w http.ResponseWriter, r *http.Request) {

	var nowUser string
	var expiresAt time.Time
	expirseHours := 24 //defa 24 max 2160(30day)

	cookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	token := cookie.Value

	err = db.DB.QueryRow("SELECT UserID, ExpiresAt FROM sessions WHERE Token = ?", token).Scan(&nowUser, &expiresAt)

	if err == sql.ErrNoRows {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if time.Now().After(expiresAt) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
}
