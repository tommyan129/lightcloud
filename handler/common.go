package handler

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"lightcloud/db"
	"log"
	"net/http"
	"syscall"
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

func GetMe(w http.ResponseWriter, r *http.Request) {
	userID, err := getSessionUser(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var username, role string
	if err = db.DB.QueryRow("SELECT Username, Role FROM users WHERE ID = ?", userID).Scan(&username, &role); err != nil {
		http.Error(w, "user not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": userID, "username": username, "role": role})
}

func GetSettings(w http.ResponseWriter, r *http.Request) {
	_, err := getSessionUser(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var name string
	err = db.DB.QueryRow("SELECT Value FROM server_settings WHERE Key = 'server_name'").Scan(&name)
	if err != nil {
		http.Error(w, "failed to get settings", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"server_name": name})
}

func GetDiskInfo(w http.ResponseWriter, r *http.Request) {
	_, err := getSessionUser(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var stat syscall.Statfs_t // wsl 환경에선 작동함 문제 없음 window 기준이라 오류 뜨는것
	err = syscall.Statfs("upload", &stat)
	if err != nil {
		log.Printf("[GetDiskInfo] Statfs: %v", err)
		http.Error(w, "failed to get disk info", http.StatusInternalServerError)
		return
	}
	bsize := uint64(stat.Bsize)
	total := stat.Blocks * bsize
	free := stat.Bavail * bsize
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]uint64{"total": total, "free": free})
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
