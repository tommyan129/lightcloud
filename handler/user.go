package handler

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"lightcloud/db"
	"lightcloud/model"

	"golang.org/x/crypto/bcrypt"
)

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
		http.Error(w, "failed to decode request", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "failed to validate request", http.StatusBadRequest)
		return
	}

	err = db.DB.QueryRow("SELECT ID FROM users WHERE Username = ?", req.Username).Scan(&existingID)
	if err == nil {
		http.Error(w, "failed to register: username already exists", http.StatusConflict)
		return
	}

	if err == sql.ErrNoRows {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("[Register] bcrypt: %v", err)
			http.Error(w, "failed to hash password", http.StatusInternalServerError)
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
			log.Printf("[Register] users INSERT: %v", err)
			http.Error(w, "failed to register user", http.StatusInternalServerError)
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
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "failed to decode request", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "failed to validate request", http.StatusBadRequest)
		return
	}

	var user model.User

	err = db.DB.QueryRow(
		"SELECT ID, PasswordHash FROM users WHERE Username = ?", req.Username,
	).Scan(&user.ID, &user.PasswordHash)
	if err == sql.ErrNoRows {
		http.Error(w, "failed to authenticate", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		http.Error(w, "failed to authenticate", http.StatusUnauthorized)
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
		log.Printf("[Login] sessions INSERT: %v", err)
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   false, // TODO: 배포 시 true로 변경 (HTTPS 필요)
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)

}

func Logout(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "failed to authenticate", http.StatusUnauthorized)
		return
	}
	token := cookie.Value

	_, err = db.DB.Exec("DELETE FROM sessions WHERE Token = ?", token)
	if err != nil {
		log.Printf("[Logout] sessions DELETE: %v", err)
		w.WriteHeader(http.StatusOK)
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

func SearchUsers(w http.ResponseWriter, r *http.Request) {
	_, err := getSessionUser(r)
	if err != nil {
		http.Error(w, "failed to authenticate", http.StatusUnauthorized)
		return
	}

	q := r.URL.Query().Get("q")
	w.Header().Set("Content-Type", "application/json")
	if q == "" {
		w.Write([]byte("[]"))
		return
	}

	rows, err := db.DB.Query("SELECT ID, Username FROM users WHERE Username LIKE ? LIMIT 10", "%"+q+"%")
	if err != nil {
		log.Printf("[SearchUsers] query: %v", err)
		http.Error(w, "failed to search users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type userResult struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	}
	var users []userResult
	for rows.Next() {
		var u userResult
		rows.Scan(&u.ID, &u.Username)
		users = append(users, u)
	}
	if users == nil {
		users = []userResult{}
	}
	json.NewEncoder(w).Encode(users)
}
