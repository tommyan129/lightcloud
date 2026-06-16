package handler

import (
	"encoding/json"
	"lightcloud/db"
	"lightcloud/model"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func Init(w http.ResponseWriter, r *http.Request) {

	existAdmin, err := AdminExists()
	if err != nil {
		log.Printf("[Init] AdminExists: %v", err)
		http.Error(w, "failed to check admin", http.StatusInternalServerError)
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
			http.Error(w, "failed to access: admin already exists", http.StatusForbidden)
			return
		}

		var req struct {
			ServerName string `json:"server_name"`
			Username   string `json:"username"`
			Password   string `json:"password"`
		}

		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "failed to decode request", http.StatusBadRequest)
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("[Init] bcrypt: %v", err)
			http.Error(w, "failed to hash password", http.StatusInternalServerError)
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
			log.Printf("[Init] begin tx: %v", err)
			http.Error(w, "failed to start tx", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		_, err = tx.Exec("INSERT INTO users (ID, Username, Role, PasswordHash, CreatedAt) VALUES (?, ?, ?, ?, ?)", user.ID, user.Username, user.Role, user.PasswordHash, user.CreatedAt.Format(time.RFC3339))
		if err != nil {
			log.Printf("[Init] users INSERT: %v", err)
			http.Error(w, "failed to create admin user", http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec("INSERT INTO server_settings (Key, Value) VALUES (?, ?)", "server_name", req.ServerName)
		if err != nil {
			log.Printf("[Init] server_settings INSERT: %v", err)
			http.Error(w, "failed to save server name", http.StatusInternalServerError)
			return
		}

		err = tx.Commit()
		if err != nil {
			log.Printf("[Init] commit tx: %v", err)
			http.Error(w, "failed to commit transaction", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusFound)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
