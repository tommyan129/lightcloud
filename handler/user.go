package handler

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"lightcloud/db"
	"lightcloud/model"

	"golang.org/x/crypto/bcrypt"
)

func generateID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatalf("ID 생성 실패: %v\n", err)
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func Register(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		http.ServeFile(w, r, "static/register.html")
		return
	}

	var existingID string
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	err = db.DB.QueryRow("SELECT ID FROM users WHERE Username = ?", req.Username).Scan(&existingID)
	if err == nil {
		w.WriteHeader(http.StatusConflict) // 409
		return
	}

	if err == sql.ErrNoRows {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		user := model.User{
			ID:           generateID(),
			Username:     req.Username,
			Role:         "user",
			PasswordHash: string(hash),
			CreatedAt:    time.Now(),
		}

		_, err = db.DB.Exec(
			"INSERT INTO users (ID, Username, Role, PasswordHash, CreatedAt) VALUES (?, ?, ?, ?, ?)", user.ID, user.Username, user.Role, user.PasswordHash, user.CreatedAt.Format(time.RFC3339),
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "회원가입 성공"})
	}
}

func Login(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		http.ServeFile(w, r, "static/login.html")
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Username == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	var user model.User

	err := db.DB.QueryRow(
		"SELECT ID, PasswordHash FROM users WHERE Username = ?", req.Username,
	).Scan(&user.ID, &user.PasswordHash)
	if err == sql.ErrNoRows {
		w.WriteHeader(http.StatusUnauthorized) // 401
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized) // 401 - 비번 틀림
		return
	}

	token := generateID()

	session := model.Session{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(2 * time.Hour),
		CreatedAt: time.Now(),
	}

	_, err = db.DB.Exec(
		"INSERT INTO sessions (Token, UserID, ExpiresAt, CreatedAt) VALUES (?, ?, ?, ?)",
		session.Token, session.UserID,
		session.ExpiresAt.Format(time.RFC3339),
		session.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)

}

func Logout(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("session")
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	token := cookie.Value

	_, err = db.DB.Exec("DELETE FROM sessions WHERE Token = ?", token)
	if err != nil {
		w.WriteHeader(http.StatusOK) //로그 아웃이니 세션 없으니 ㅇㅋ 니 로그인으로 돌아가라 시전
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   "",
		Expires: time.Unix(0, 0),
		Path:    "/",
	})

	w.WriteHeader(http.StatusOK)
}
