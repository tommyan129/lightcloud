package handler

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"lightcloud/db"
	"lightcloud/model"

	"golang.org/x/crypto/bcrypt"
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

	var req struct {
		FileIDs      []string `json:"file_ids"`
		ExpiresHours int      `json:"expires_hours"`
		Password     string   `json:"password"`
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) //400
		return
	}

	if req.ExpiresHours < 24 {
		req.ExpiresHours = 24
	}

	if req.ExpiresHours > 2160 {
		req.ExpiresHours = 2160
	}

	var hash []byte

	if req.Password != "" {
		hash, err = bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	sharelink := model.ShareLink{
		Token:        generateShareToken(),
		CreatedBy:    nowUser,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Duration(req.ExpiresHours) * time.Hour),
		PasswordHash: string(hash),
	}

	tx, err := db.DB.Begin()
	if err != nil {
		//err
		return
	}

	_, err = tx.Exec("INSERT INTO share_links (Token, CreatedAt, CreatedBy, ExpiresAt, PasswordHash) VALUES (?, ?, ?, ?, ?)", sharelink.Token, sharelink.CreatedAt.Format(time.RFC3339), sharelink.CreatedBy, sharelink.ExpiresAt.Format(time.RFC3339), sharelink.PasswordHash)
	if err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, id := range req.FileIDs {
		var p int

		err = db.DB.QueryRow("SELECT Permission FROM file_permissions WHERE UserID = ? AND FileID = ?", nowUser, id).Scan(&p)
		if err == sql.ErrNoRows {
			log.Printf("share: permission denied user=%s file=%s", nowUser, id)
			continue
		}
		if (p & model.PermDownload) == 0 {
			log.Printf("share: permission denied user=%s file=%s", nowUser, id)
			continue
		}

		_, err = tx.Exec("INSERT INTO share_link_files (Token, FileID) VALUES (?, ?)", sharelink.Token, id)
		if err != nil {
			tx.Rollback()
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	tx.Commit()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		Token string `json:"token"`
	}{sharelink.Token})
}

func getLinkFiles(token string, w http.ResponseWriter, r *http.Request) (files []model.File, err error) {
	rows, err := db.DB.Query("SELECT f.ID, f.OriginalName, f.Size, f.MimeType  FROM files f JOIN share_link_files s ON f.ID = s.FileID WHERE s.Token = ?", token)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var f model.File
		err = rows.Scan(&f.ID, &f.OriginalName, &f.Size, &f.MimeType)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		files = append(files, f)
	}

	return files, err
}

func GetShareLink(w http.ResponseWriter, r *http.Request) {

	var sharelink model.ShareLink

	token := r.URL.Query().Get("token")
	if token == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := db.DB.QueryRow("SELECT ExpiresAt, PasswordHash From share_links WHERE Token = ?", token).Scan(&sharelink.ExpiresAt, &sharelink.PasswordHash)

	if err == sql.ErrNoRows {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if time.Now().After(sharelink.ExpiresAt) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if sharelink.PasswordHash != "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(struct {
				RequirePassword bool `json:"requires_password"`
			}{true})
		}
		if sharelink.PasswordHash == "" {

			files, err := getLinkFiles(token, w, r)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			data, err := json.Marshal(&files)
			if err != nil {
				log.Printf("JSON 직렬화 실패: %v", err)
				http.Error(w, "json encode failed", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		}
		break

	case http.MethodPost:
		var req struct {
			Password string `json:"password"`
		}
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if req.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(sharelink.PasswordHash), []byte(req.Password))
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized) // 401 - 비번 틀림
			return
		}

		files, err := getLinkFiles(token, w, r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		data, err := json.Marshal(&files)
		if err != nil {
			log.Printf("JSON 직렬화 실패: %v", err)
			http.Error(w, "json encode failed", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)

		break
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
