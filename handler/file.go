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

	var nowUser string
	var expiresAt time.Time
	var files []model.File

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

	rows, err := db.DB.Query("SELECT ID, OriginalName, Size FROM files WHERE OwnerID = ?", nowUser)
	if err != nil {
		log.Printf("파일 목록 조회 실패: %v", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var f model.File
		err = rows.Scan(&f.ID, &f.OriginalName, &f.Size)
		if err != nil {
			log.Printf("파일 행 읽기 실패: %v", err)
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		files = append(files, f)
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

func UploadFiles(w http.ResponseWriter, r *http.Request) {

	var nowUser string
	var expiresAt time.Time
	var adminID string

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

	r.ParseMultipartForm(32 << 20)

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
			CreatedAt:    time.Now(),
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

		_, err = db.DB.Exec("INSERT INTO files (ID, OwnerID, OriginalName, StoredName, Size, MimeType, CreatedAt) VALUES (?, ?, ?, ?, ?, ?, ?)", newFile.ID, newFile.OwnerID, newFile.OriginalName, newFile.StoredName, newFile.Size, newFile.MimeType, newFile.CreatedAt.Format(time.RFC3339))
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

		if adminID != "" {
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
}

func DownloadFiles(w http.ResponseWriter, r *http.Request) {
	var nowUser string
	var expiresAt time.Time
	var file model.File
	var p int //perm 받는 임시 함수

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

func MainPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/main.html")
}
