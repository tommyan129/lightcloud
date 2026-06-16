package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"lightcloud/db"
	"lightcloud/model"
)

func requireAdmin(r *http.Request) (string, error) {
	nowUser, err := getSessionUser(r)
	if err != nil {
		return "", err
	}
	var role string
	if err = db.DB.QueryRow("SELECT Role FROM users WHERE ID = ?", nowUser).Scan(&role); err != nil {
		return "", err
	}
	if role != "admin" && role != "assiadmin" {
		return "", fmt.Errorf("forbidden")
	}
	return nowUser, nil
}

func AdminPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/admin.html")
}

func GetAdminStats(w http.ResponseWriter, r *http.Request) {
	if _, err := requireAdmin(r); err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	var userCount, fileCount, shareLinkCount int
	var totalSize int64
	db.DB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	db.DB.QueryRow("SELECT COUNT(*), COALESCE(SUM(Size), 0) FROM files").Scan(&fileCount, &totalSize)
	db.DB.QueryRow("SELECT COUNT(*) FROM share_links").Scan(&shareLinkCount)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_count":       userCount,
		"file_count":       fileCount,
		"total_size":       totalSize,
		"share_link_count": shareLinkCount,
	})
}

func GetAdminUsers(w http.ResponseWriter, r *http.Request) {
	if _, err := requireAdmin(r); err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	rows, err := db.DB.Query("SELECT ID, Username, Role, CreatedAt FROM users ORDER BY CreatedAt")
	if err != nil {
		log.Printf("[GetAdminUsers] query: %v", err)
		http.Error(w, "failed to query users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	users := []model.User{}
	for rows.Next() {
		var u model.User
		var createdAt string
		rows.Scan(&u.ID, &u.Username, &u.Role, &createdAt)
		u.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		users = append(users, u)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	nowUser, err := requireAdmin(r)
	if err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	var req struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.Role != "user" && req.Role != "assiadmin" {
		http.Error(w, "invalid role", http.StatusBadRequest)
		return
	}
	if req.UserID == nowUser {
		http.Error(w, "cannot change own role", http.StatusForbidden)
		return
	}
	var targetRole string
	if err = db.DB.QueryRow("SELECT Role FROM users WHERE ID = ?", req.UserID).Scan(&targetRole); err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	if targetRole == "admin" {
		http.Error(w, "cannot change admin role", http.StatusForbidden)
		return
	}
	if _, err = db.DB.Exec("UPDATE users SET Role = ? WHERE ID = ?", req.Role, req.UserID); err != nil {
		log.Printf("[UpdateUserRole] UPDATE [%s]: %v", req.UserID, err)
		http.Error(w, "failed to update role", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

func DeleteAdminUser(w http.ResponseWriter, r *http.Request) {
	nowUser, err := requireAdmin(r)
	if err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	var req struct {
		UserID string `json:"user_id"`
	}
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if req.UserID == nowUser {
		http.Error(w, "cannot delete yourself", http.StatusForbidden)
		return
	}
	var targetRole string
	if err = db.DB.QueryRow("SELECT Role FROM users WHERE ID = ?", req.UserID).Scan(&targetRole); err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	if targetRole == "admin" {
		http.Error(w, "cannot delete admin", http.StatusForbidden)
		return
	}
	if _, err = db.DB.Exec("DELETE FROM users WHERE ID = ?", req.UserID); err != nil {
		log.Printf("[DeleteAdminUser] DELETE [%s]: %v", req.UserID, err)
		http.Error(w, "failed to delete user", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

func GetAdminFiles(w http.ResponseWriter, r *http.Request) {
	if _, err := requireAdmin(r); err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	rows, err := db.DB.Query(`
		SELECT f.ID, COALESCE(f.FolderID,''), f.OriginalName, f.Size, f.MimeType, f.UploadedAt, f.OwnerID, u.Username
		FROM files f JOIN users u ON f.OwnerID = u.ID
		ORDER BY u.Username, f.UploadedAt DESC`)
	if err != nil {
		log.Printf("[GetAdminFiles] query: %v", err)
		http.Error(w, "failed to query files", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	files := []model.File{}
	for rows.Next() {
		var f model.File
		rows.Scan(&f.ID, &f.FolderId, &f.OriginalName, &f.Size, &f.MimeType, &f.UploadedAt, &f.OwnerID, &f.OwnerName)
		files = append(files, f)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func DeleteAdminFile(w http.ResponseWriter, r *http.Request) {
	if _, err := requireAdmin(r); err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	var storedName string
	if err := db.DB.QueryRow("SELECT StoredName FROM files WHERE ID = ?", req.ID).Scan(&storedName); err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	os.Remove(filepath.Join(uploadFilesPath, storedName))
	if _, err := db.DB.Exec("DELETE FROM files WHERE ID = ?", req.ID); err != nil {
		log.Printf("[DeleteAdminFile] DELETE [%s]: %v", req.ID, err)
		http.Error(w, "failed to delete file", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

func GetAdminShares(w http.ResponseWriter, r *http.Request) {
	if _, err := requireAdmin(r); err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	rows, err := db.DB.Query(`
		SELECT sl.Token, COALESCE(sl.Title,''), sl.CreatedBy, u.Username,
		       sl.CreatedAt, sl.ExpiresAt,
		       CASE WHEN sl.PasswordHash IS NOT NULL THEN 1 ELSE 0 END,
		       COUNT(slf.FileID)
		FROM share_links sl
		JOIN users u ON sl.CreatedBy = u.ID
		LEFT JOIN share_link_files slf ON sl.Token = slf.Token
		GROUP BY sl.Token
		ORDER BY sl.CreatedAt DESC`)
	if err != nil {
		log.Printf("[GetAdminShares] query: %v", err)
		http.Error(w, "failed to query shares", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	type ShareRow struct {
		Token       string `json:"token"`
		Title       string `json:"title"`
		CreatorName string `json:"creator_name"`
		CreatedAt   string `json:"created_at"`
		ExpiresAt   string `json:"expires_at"`
		HasPassword bool   `json:"has_password"`
		FileCount   int    `json:"file_count"`
	}
	shares := []ShareRow{}
	for rows.Next() {
		var s ShareRow
		var createdBy string
		var hp int
		rows.Scan(&s.Token, &s.Title, &createdBy, &s.CreatorName, &s.CreatedAt, &s.ExpiresAt, &hp, &s.FileCount)
		s.HasPassword = hp == 1
		shares = append(shares, s)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(shares)
}

func DeleteAdminShare(w http.ResponseWriter, r *http.Request) {
	if _, err := requireAdmin(r); err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if _, err := db.DB.Exec("DELETE FROM share_links WHERE Token = ?", req.Token); err != nil {
		log.Printf("[DeleteAdminShare] DELETE [%s]: %v", req.Token, err)
		http.Error(w, "failed to delete share link", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

func GetAdminSessions(w http.ResponseWriter, r *http.Request) {
	if _, err := requireAdmin(r); err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	rows, err := db.DB.Query(`
		SELECT s.Token, s.UserID, u.Username, s.CreatedAt, s.ExpiresAt
		FROM sessions s
		JOIN users u ON s.UserID = u.ID
		ORDER BY s.ExpiresAt DESC`)
	if err != nil {
		log.Printf("[GetAdminSessions] query: %v", err)
		http.Error(w, "failed to query sessions", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	type SessionRow struct {
		Token     string `json:"token"`
		UserID    string `json:"user_id"`
		Username  string `json:"username"`
		CreatedAt string `json:"created_at"`
		ExpiresAt string `json:"expires_at"`
	}
	sessions := []SessionRow{}
	for rows.Next() {
		var s SessionRow
		rows.Scan(&s.Token, &s.UserID, &s.Username, &s.CreatedAt, &s.ExpiresAt)
		sessions = append(sessions, s)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

func DeleteAdminSession(w http.ResponseWriter, r *http.Request) {
	if _, err := requireAdmin(r); err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if _, err := db.DB.Exec("DELETE FROM sessions WHERE Token = ?", req.Token); err != nil {
		log.Printf("[DeleteAdminSession] DELETE: %v", err)
		http.Error(w, "failed to delete session", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}
