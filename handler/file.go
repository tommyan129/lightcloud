package handler

import (
	"archive/zip"
	"database/sql"
	"encoding/json"
	"strconv"

	"io"
	"log"
	"net/http"
	"time"

	"os"
	"path/filepath"

	"lightcloud/db"
	"lightcloud/model"
)

const uploadFilesPath = "./upload"

func ListFiles(w http.ResponseWriter, r *http.Request) {

	response := model.FileListResponse{
		Mine:   []model.File{},
		Shared: []model.File{},
	}

	nowUser, err := getSessionUser(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	rows, err := db.DB.Query("SELECT ID, OriginalName, Size, MimeType, UploadedAt FROM files WHERE OwnerID = ?", nowUser)
	if err != nil {
		log.Printf("파일 목록 조회 실패: %v", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var f model.File
		err = rows.Scan(&f.ID, &f.OriginalName, &f.Size, &f.MimeType, &f.UploadedAt)
		if err != nil {
			log.Printf("파일 행 읽기 실패: %v", err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		response.Mine = append(response.Mine, f)
	}
	rows.Close()

	rows, err = db.DB.Query(`SELECT f.ID, f.OriginalName, f.Size, f.MimeType, f.UploadedAt, u.Username FROM files f JOIN file_permissions fp ON f.ID = fp.FileID JOIN users u ON f.OwnerID = u.ID WHERE fp.UserID = ? AND f.OwnerID != ? ORDER BY u.Username`, nowUser, nowUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var f model.File
		err = rows.Scan(&f.ID, &f.OriginalName, &f.Size, &f.MimeType, &f.UploadedAt, &f.OwnerName)
		if err != nil {
			log.Printf("파일 행 읽기 실패: %v", err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		response.Shared = append(response.Shared, f)
	}
	rows.Close()

	data, err := json.Marshal(&response)
	if err != nil {
		log.Printf("JSON 직렬화 실패: %v", err)
		http.Error(w, "json encode failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func UploadFiles(w http.ResponseWriter, r *http.Request) {

	var adminID string

	nowUser, err := getSessionUser(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err = r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "invalid multipart form", http.StatusBadRequest)
		return
	}

	fileHeaders := r.MultipartForm.File["file"]
	for _, fileHeader := range fileHeaders {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "file open failed", http.StatusInternalServerError)
			return
		}

		newFile := model.File{
			ID:           generateID(),
			OwnerID:      nowUser,
			OriginalName: fileHeader.Filename,
			Size:         fileHeader.Size,
			MimeType:     fileHeader.Header.Get("Content-Type"),
			UploadedAt:   time.Now().Format(time.RFC3339),
		}

		ownerPerm := model.FilePermission{
			FileID:     newFile.ID,
			UserID:     nowUser,
			Permission: model.PermRead | model.PermDownload | model.PermWrite | model.PermDelete | model.PermManage,
		}

		err = db.DB.QueryRow("SELECT ID FROM users WHERE Role = 'admin'").Scan(&adminID)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("admin 조회 실패: %v", err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}

		adminPerm := model.FilePermission{
			FileID:     newFile.ID,
			UserID:     adminID,
			Permission: model.PermRead | model.PermDownload | model.PermWrite | model.PermDelete | model.PermManage,
		}

		assistAdminRows, err := db.DB.Query("SELECT ID FROM users WHERE Role = 'assiadmin'")
		if err != nil {
			log.Printf("assiadmin 조회 실패: %v", err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}

		var assiadminPerms []model.FilePermission

		for assistAdminRows.Next() {
			var assiadminID string
			assistAdminRows.Scan(&assiadminID)
			assiadminPerm := model.FilePermission{
				FileID:     newFile.ID,
				UserID:     assiadminID,
				Permission: model.PermRead | model.PermDownload | model.PermWrite | model.PermDelete | model.PermManage,
			}
			assiadminPerms = append(assiadminPerms, assiadminPerm)
		}
		assistAdminRows.Close()

		newFile.StoredName = newFile.ID + filepath.Ext(fileHeader.Filename)

		savedFile, err := os.Create(filepath.Join(uploadFilesPath, newFile.StoredName))
		if err != nil {
			log.Printf("파일 생성 실패 [%s]: %v", newFile.StoredName, err)
			http.Error(w, "file save failed", http.StatusInternalServerError)
			return
		}

		_, err = io.Copy(savedFile, file)
		if err != nil {
			log.Printf("파일 복사 실패 [%s]: %v", newFile.StoredName, err)
			http.Error(w, "file write failed", http.StatusInternalServerError)
			return
		} //업로드 된 용량 실시간 전달해서 얼마나 올라갔는지 보이게 하는기능
		file.Close()
		savedFile.Close()

		_, err = db.DB.Exec("INSERT INTO files (ID, OwnerID, OriginalName, StoredName, Size, MimeType, UploadedAt) VALUES (?, ?, ?, ?, ?, ?, ?)", newFile.ID, newFile.OwnerID, newFile.OriginalName, newFile.StoredName, newFile.Size, newFile.MimeType, newFile.UploadedAt)
		if err != nil {
			log.Printf("files INSERT 실패 [%s]: %v", newFile.ID, err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}

		_, err = db.DB.Exec("INSERT INTO file_permissions (FileID, UserID, Permission) VALUES (?, ?, ?)", ownerPerm.FileID, ownerPerm.UserID, ownerPerm.Permission)
		if err != nil {
			log.Printf("ownerPerm INSERT 실패 [%s]: %v", ownerPerm.FileID, err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}

		if adminID != "" && nowUser != adminID {
			_, err = db.DB.Exec("INSERT INTO file_permissions (FileID, UserID, Permission) VALUES (?, ?, ?)", adminPerm.FileID, adminPerm.UserID, adminPerm.Permission)
			if err != nil {
				log.Printf("adminPerm INSERT 실패 [%s]: %v", adminPerm.FileID, err)
				http.Error(w, "db error", http.StatusInternalServerError)
				return
			}
		}

		for _, p := range assiadminPerms {
			_, err = db.DB.Exec("INSERT INTO file_permissions (FileID, UserID, Permission) VALUES (?, ?, ?)", p.FileID, p.UserID, p.Permission)
			if err != nil {
				log.Printf("assiadminPerm INSERT 실패 [%s]: %v", p.FileID, err)
				http.Error(w, "db error", http.StatusInternalServerError)
				return
			}
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

func DownloadFiles(w http.ResponseWriter, r *http.Request) {
	var file model.File
	var p int //perm 받는 임시 함수

	nowUser, err := getSessionUser(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	getFileIDs := r.URL.Query()["id"]
	if len(getFileIDs) == 0 {
		http.Error(w, "emptyIDs", http.StatusBadRequest)
		return
	}

	if len(getFileIDs) == 1 {
		file.ID = getFileIDs[0]
		err = db.DB.QueryRow("SELECT OwnerID, OriginalName, StoredName, Size, MimeType FROM files WHERE ID = ?", file.ID).Scan(&file.OwnerID, &file.OriginalName, &file.StoredName, &file.Size, &file.MimeType)
		if err != nil {
			http.Error(w, "cant find file", http.StatusNotFound)
			return
		}

		err = db.DB.QueryRow("SELECT Permission From file_permissions WHERE FileID = ? AND UserID = ?", file.ID, nowUser).Scan(&p)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if (p & model.PermDownload) == 0 {
			http.Error(w, "unauthorized", http.StatusForbidden)
			return
		}

		getFile, err := os.Open(filepath.Join(uploadFilesPath, file.StoredName))
		if err != nil {
			log.Printf("파일 열기 실패 [%s]: %v", file.StoredName, err)
			http.Error(w, "file open failed", http.StatusInternalServerError)
			return
		}
		defer getFile.Close()

		w.Header().Set("Content-Type", file.MimeType)
		w.Header().Set("Content-Disposition", "attachment; filename=\""+file.OriginalName+"\"")

		_, err = io.Copy(w, getFile) //여기도 다운로드 현황 브라우져에서 자동으로 띄워주니 상관없나?
		if err != nil {
			log.Printf("파일 전송 실패 [%s]: %v", file.StoredName, err)
			http.Error(w, "file send failed", http.StatusInternalServerError)
			return
		}

	}

	if len(getFileIDs) >= 2 {
		zipName := time.Now().Format("200601021504") + "_" + strconv.Itoa(len(getFileIDs)) + ".zip"
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+zipName+"\"")

		zipWriter := zip.NewWriter(w)

		for _, id := range getFileIDs {
			err = db.DB.QueryRow("SELECT OwnerID, OriginalName, StoredName, Size, MimeType FROM files WHERE ID = ?", id).Scan(&file.OwnerID, &file.OriginalName, &file.StoredName, &file.Size, &file.MimeType)
			if err != nil {
				http.Error(w, "cant find file", http.StatusNotFound)
				return
			}

			err = db.DB.QueryRow("SELECT Permission From file_permissions WHERE FileID = ? AND UserID = ?", id, nowUser).Scan(&p)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if (p & model.PermDownload) == 0 {
				log.Printf("다운로드 권한 없음 [%s]: skip", id)
				continue
			}

			getFile, err := os.Open(filepath.Join(uploadFilesPath, file.StoredName))
			if err != nil {
				log.Printf("파일 열기 실패 [%s]: %v", file.StoredName, err)
				http.Error(w, "file open failed", http.StatusInternalServerError)
				return
			}

			entry, err := zipWriter.Create(file.OriginalName)
			if err != nil {
				log.Printf("zip 항목 생성 실패 [%s]: %v", file.OriginalName, err)
				getFile.Close()
				zipWriter.Close()
				return
			}

			_, err = io.Copy(entry, getFile) //여기도 다운로드 현황 브라우져에서 자동으로 띄워주니 상관없나?
			if err != nil {
				log.Printf("파일 전송 실패 [%s]: %v", file.StoredName, err)
				http.Error(w, "file send failed", http.StatusInternalServerError)
				getFile.Close()
				return
			}
			getFile.Close()
		}
		zipWriter.Close()
	}
}

func DeleteFiles(w http.ResponseWriter, r *http.Request) {

	var req struct {
		IDs []string `json:"ids"`
	}
	var delFiles []model.File
	var p int

	nowUser, err := getSessionUser(r)
	if err != nil {
		http.Error(w, "failed getsession", http.StatusInternalServerError)
		return
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "failed decode json", http.StatusInternalServerError)
		return
	}

	for _, id := range req.IDs {
		var f model.File
		f.ID = id
		err = db.DB.QueryRow("SELECT OwnerID, StoredName FROM files WHERE ID = ?", id).Scan(&f.OwnerID, &f.StoredName)
		if err != nil {
			http.Error(w, "failed to find data", http.StatusInternalServerError)
			return
		}

		err = db.DB.QueryRow("SELECT Permission FROM file_permissions WHERE FileID = ? AND UserID = ?", id, nowUser).Scan(&p)
		if err != nil {
			http.Error(w, "failed to find data", http.StatusInternalServerError)
			return
		}

		if (p & model.PermDelete) == 0 {
			log.Printf("삭제 권한 없음 [%s]: skip", id)
			continue
		}
		delFiles = append(delFiles, f)
	}

	for _, del := range delFiles {
		err = os.Remove(filepath.Join(uploadFilesPath, del.StoredName))
		if err != nil {
			http.Error(w, "파일 삭제 실패", http.StatusInternalServerError)
			return
		}
		_, err = db.DB.Exec("DELETE FROM files WHERE ID = ?", del.ID)
		if err != nil {
			log.Printf("파일 DB 삭제 실패 [%s]: %v", del.ID, err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

func GetGrantedPerms(w http.ResponseWriter, r *http.Request) {
	nowUser, err := getSessionUser(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	type PermItem struct {
		FileID     string `json:"file_id"`
		FileName   string `json:"file_name"`
		UserID     string `json:"user_id"`
		Username   string `json:"username"`
		Permission int    `json:"permission"`
	}

	rows, err := db.DB.Query(`
		SELECT f.ID, f.OriginalName, fp.UserID, u.Username, fp.Permission
		FROM files f
		JOIN file_permissions fp ON f.ID = fp.FileID
		JOIN users u ON fp.UserID = u.ID
		WHERE f.OwnerID = ? AND fp.UserID != ?
		ORDER BY f.OriginalName, u.Username`, nowUser, nowUser)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []PermItem
	for rows.Next() {
		var item PermItem
		rows.Scan(&item.FileID, &item.FileName, &item.UserID, &item.Username, &item.Permission)
		items = append(items, item)
	}
	if items == nil {
		items = []PermItem{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func UpdatePerm(w http.ResponseWriter, r *http.Request) {
	var p int

	nowUser, err := getSessionUser(r)
	if err != nil {
		http.Error(w, "failed getsession", http.StatusInternalServerError)
		return
	}
	var req struct {
		FileID      string `json:"file_id"`
		Permissions []struct {
			UserID     string `json:"user_id"`
			Permission int    `json:"permission"`
		} `json:"permissions"`
	}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "failed decode json", http.StatusInternalServerError)
		return
	}

	err = db.DB.QueryRow("SELECT permission FROM file_permissions WHERE FileID =? AND UserID = ?", req.FileID, nowUser).Scan(&p)
	if err != nil {
		http.Error(w, "failed to find data", http.StatusInternalServerError)
		return
	}

	if (p & model.PermManage) == 0 {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	for _, perm := range req.Permissions {
		_, err = db.DB.Exec("INSERT INTO file_permissions (FileID, UserID, Permission) VALUES (?, ?, ?) ON CONFLICT (FileID, UserID) DO UPDATE SET Permission = excluded.Permission", req.FileID, perm.UserID, perm.Permission)
		if err != nil {
			http.Error(w, "failed to update perm", http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

func UpdateOwner(w http.ResponseWriter, r *http.Request) {

	var req struct {
		FileID     string `json:"file_id"`
		TargetUser string `json:"targetuser"`
	}
	var p int
	var ro string

	nowUser, err := getSessionUser(r)
	if err != nil {
		http.Error(w, "failed getsession", http.StatusInternalServerError)
		return
	}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "failed decode json", http.StatusInternalServerError)
		return
	}

	err = db.DB.QueryRow("SELECT Permission FROM file_permissions WHERE FileID = ? AND UserID = ?", req.FileID, nowUser).Scan(&p)
	if err != nil {
		http.Error(w, "failed to find data", http.StatusInternalServerError)
		return
	}

	if (p & model.PermManage) == 0 {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	_, err = db.DB.Exec("UPDATE files SET OwnerID = ? WHERE ID = ?", req.TargetUser, req.FileID)
	if err != nil {
		//err
		return
	}
	_, err = db.DB.Exec("INSERT INTO file_permissions (FileID, UserID, Permission) VALUES (?, ?, ?) ON CONFLICT (FileID, UserID) DO UPDATE SET Permission = excluded.Permission", req.FileID, req.TargetUser, model.PermRead|model.PermDownload|model.PermWrite|model.PermDelete|
		model.PermManage)
	if err != nil {
		http.Error(w, "failed to update perm", http.StatusInternalServerError)
		return
	}

	err = db.DB.QueryRow("SELECT role FROM users WHERE ID = ?", nowUser).Scan(&ro)
	if err != nil {
		http.Error(w, "failed to find data", http.StatusInternalServerError)
		return
	}

	if ro != "admin" && ro != "assiadmin" {
		_, err = db.DB.Exec("INSERT INTO file_permissions (FileID, UserID, Permission) VALUES (?, ?, ?) ON CONFLICT (FileID, UserID) DO UPDATE SET Permission = excluded.Permission", req.FileID, nowUser, model.PermRead|model.PermDownload)
		if err != nil {
			http.Error(w, "failed to update perm", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}
