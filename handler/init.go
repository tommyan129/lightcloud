package handler

import (
	"encoding/json"
	"lightcloud/db"
	"lightcloud/model"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func Init(w http.ResponseWriter, r *http.Request) {

	existAdmin, err := AdminExists()
	if err != nil {
		http.Error(w, "failed to find db res", http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if existAdmin == false {
			http.ServeFile(w, r, "static/init.html")
			return
		}
		http.Redirect(w, r, "/login", http.StatusFound)

	case http.MethodPost:
		if existAdmin == true {
			http.Error(w, "unauthorized", http.StatusForbidden)
			return
		}

		var req struct {
			ServerName string `json:"server_name"`
			Username   string `json:"username"`
			Password   string `json:"password"`
		}

		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "failed to pasing to json", http.StatusInternalServerError)
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "failed to make pwhash", http.StatusInternalServerError)
			return
		}

		user := model.User{
			ID:           generateID(),
			Username:     req.Username,
			Role:         "admin",
			PasswordHash: string(hash),
			CreatedAt:    time.Now(),
		}
		tx, err := db.DB.Begin()
		if err != nil {
			http.Error(w, "failed to start tx", http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec("INSERT INTO users (ID, Username, Role, PasswordHash, CreatedAt) VALUES (?, ?, ?, ?, ?)", user.ID, user.Username, user.Role, user.PasswordHash, user.CreatedAt.Format(time.RFC3339))
		if err != nil {
			tx.Rollback()
			http.Error(w, "failed to insert admin role", http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec("INSERT INTO server_settings (Key, Value) VALUES (?, ?)", "server_name", req.ServerName)
		if err != nil {
			tx.Rollback()
			http.Error(w, "failed to insert server_name", http.StatusInternalServerError)
			return
		}

		tx.Commit()

		http.Redirect(w, r, "/login", http.StatusFound)

	default:
		http.Error(w, "not allowed method", http.StatusMethodNotAllowed)
	}
}
