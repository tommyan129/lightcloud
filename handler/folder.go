package handler

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"lightcloud/db"
	"lightcloud/model"
)

func ListFolders(w http.ResponseWriter, r *http.Request) {
	nowUser, err := getSessionUser(r)
	if err != nil {
		http.Error(w, "failed to authenticate", http.StatusUnauthorized)
		return
	}

	rows, err := db.DB.Query("SELECT ID, ParentID, Name, CreatedAt FROM folders WHERE OwnerID = ? ORDER BY Name", nowUser)
	if err != nil {
		log.Printf("[ListFolders] query: %v", err)
		http.Error(w, "failed to query folders", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	folders := []model.Folder{}
	for rows.Next() {
		var f model.Folder
		var parentID sql.NullString
		var createdAtStr string
		err = rows.Scan(&f.ID, &parentID, &f.Name, &createdAtStr)
		if err != nil {
			log.Printf("[ListFolders] scan: %v", err)
			http.Error(w, "failed to read folders", http.StatusInternalServerError)
			return
		}
		f.OwnerID = nowUser
		f.ParentID = parentID.String
		f.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		folders = append(folders, f)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(folders)
}

func CreateFolder(w http.ResponseWriter, r *http.Request) {
	nowUser, err := getSessionUser(r)
	if err != nil {
		http.Error(w, "failed to authenticate", http.StatusUnauthorized)
		return
	}

	var req struct {
		Name     string `json:"name"`
		ParentID string `json:"parent_id"`
	}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "failed to decode request", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if req.ParentID != "" {
		var ownerID string
		err = db.DB.QueryRow("SELECT OwnerID FROM folders WHERE ID = ?", req.ParentID).Scan(&ownerID)
		if err != nil {
			log.Printf("[CreateFolder] parent query [%s]: %v", req.ParentID, err)
			http.Error(w, "failed to find parent folder", http.StatusNotFound)
			return
		}
		if ownerID != nowUser {
			http.Error(w, "failed to access folder", http.StatusForbidden)
			return
		}
	}

	folder := model.Folder{
		ID:        generateID(),
		OwnerID:   nowUser,
		ParentID:  req.ParentID,
		Name:      req.Name,
		CreatedAt: time.Now(),
	}

	var parentID any //nil 넣기
	if req.ParentID != "" {
		parentID = req.ParentID
	}

	_, err = db.DB.Exec("INSERT INTO folders (ID, OwnerID, ParentID, Name, CreatedAt) VALUES (?, ?, ?, ?, ?)",
		folder.ID, folder.OwnerID, parentID, folder.Name, folder.CreatedAt.Format(time.RFC3339))
	if err != nil {
		log.Printf("[CreateFolder] INSERT [%s]: %v", folder.ID, err)
		http.Error(w, "failed to create folder", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		ID string `json:"id"`
	}{folder.ID})
}

func DeleteFolder(w http.ResponseWriter, r *http.Request) {
	nowUser, err := getSessionUser(r)
	if err != nil {
		http.Error(w, "failed to authenticate", http.StatusUnauthorized)
		return
	}

	var req struct {
		FolderID string `json:"folder_id"`
	}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "failed to decode request", http.StatusBadRequest)
		return
	}

	var ownerID string
	err = db.DB.QueryRow("SELECT OwnerID FROM folders WHERE ID = ?", req.FolderID).Scan(&ownerID)
	if err != nil {
		log.Printf("[DeleteFolder] query [%s]: %v", req.FolderID, err)
		http.Error(w, "failed to find folder", http.StatusNotFound)
		return
	}

	if ownerID != nowUser {
		http.Error(w, "failed to access folder", http.StatusForbidden)
		return
	}

	_, err = db.DB.Exec("DELETE FROM folders WHERE ID = ?", req.FolderID)
	if err != nil {
		log.Printf("[DeleteFolder] DELETE [%s]: %v", req.FolderID, err)
		http.Error(w, "failed to delete folder", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

func MoveFiles(w http.ResponseWriter, r *http.Request) {
	nowUser, err := getSessionUser(r)
	if err != nil {
		http.Error(w, "failed to authenticate", http.StatusUnauthorized)
		return
	}

	var req struct {
		FileIDs  []string `json:"file_ids"`
		FolderID string   `json:"folder_id"`
	}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "failed to decode request", http.StatusBadRequest)
		return
	}

	if req.FolderID != "" {
		var ownerID string
		err = db.DB.QueryRow("SELECT OwnerID FROM folders WHERE ID = ?", req.FolderID).Scan(&ownerID)
		if err != nil {
			log.Printf("[MoveFiles] folder query [%s]: %v", req.FolderID, err)
			http.Error(w, "failed to find folder", http.StatusNotFound)
			return
		}
		if ownerID != nowUser {
			http.Error(w, "failed to access folder", http.StatusForbidden)
			return
		}
	}

	var folderID any
	if req.FolderID != "" {
		folderID = req.FolderID
	}

	for _, id := range req.FileIDs {
		var ownerID string
		err = db.DB.QueryRow("SELECT OwnerID FROM files WHERE ID = ?", id).Scan(&ownerID)
		if err != nil {
			log.Printf("[MoveFiles] file query [%s]: %v", id, err)
			continue
		}
		if ownerID != nowUser {
			log.Printf("[MoveFiles] not owner [%s]: skip", id)
			continue
		}

		_, err = db.DB.Exec("UPDATE files SET FolderID = ? WHERE ID = ?", folderID, id)
		if err != nil {
			log.Printf("[MoveFiles] UPDATE [%s]: %v", id, err)
			http.Error(w, "failed to move file", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}
