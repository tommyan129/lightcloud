package handler

import (
	"crypto/rand"
	"fmt"
	"lightcloud/db"
	"log"
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

func generateID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatalf("ID 생성 실패: %v\n", err)
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

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

func AdminExists() (bool, error) {

	var res int
	err := db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE Role = ?", "admin").Scan(&res)
	if err != nil {
		//err
		return false, err
	}

	if res > 0 {
		return true, nil
	}

	return false, nil
}
