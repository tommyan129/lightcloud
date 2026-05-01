package handler

import (
	"archive/zip"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"lightcloud/db"
	"lightcloud/model"

	"golang.org/x/crypto/bcrypt"
)

func GetMyShareLinks(w http.ResponseWriter, r *http.Request) {
	nowUser, err := getSessionUser(r)
	if err != nil {
		http.Error(w, "failed to authenticate", http.StatusUnauthorized)
		return
	}

	type ShareLinkItem struct {
		Token       string `json:"token"`
		Title       string `json:"share_title"`
		CreatedAt   string `json:"created_at"`
		ExpiresAt   string `json:"expires_at"`
		HasPassword bool   `json:"has_password"`
		FileCount   int    `json:"file_count"`
	}

	rows, err := db.DB.Query(`
		SELECT sl.Token, sl.Title, sl.CreatedAt, sl.ExpiresAt,
		       CASE WHEN sl.PasswordHash != '' AND sl.PasswordHash IS NOT NULL THEN 1 ELSE 0 END,
		       COUNT(slf.FileID)
		FROM share_links sl
		LEFT JOIN share_link_files slf ON sl.Token = slf.Token
		WHERE sl.CreatedBy = ?
		GROUP BY sl.Token
		ORDER BY sl.CreatedAt DESC`, nowUser)
	if err != nil {
		log.Printf("[GetMyShareLinks] query: %v", err)
		http.Error(w, "failed to query share links", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var items []ShareLinkItem
	for rows.Next() {
		var item ShareLinkItem
		var hasPw int
		rows.Scan(&item.Token, &item.Title, &item.CreatedAt, &item.ExpiresAt, &hasPw, &item.FileCount)
		item.HasPassword = hasPw == 1
		items = append(items, item)
	}
	if items == nil {
		items = []ShareLinkItem{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func CreateShareLink(w http.ResponseWriter, r *http.Request) {

	nowUser, err := getSessionUser(r)
	if err != nil {
		http.Error(w, "failed to authenticate", http.StatusUnauthorized)
		return
	}

	var req struct {
		FileIDs      []string `json:"file_ids"`
		ExpiresHours int      `json:"expires_hours"`
		Password     string   `json:"password"`
		Title        string   `json:"share_title"`
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "failed to decode request", http.StatusBadRequest)
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
			log.Printf("[CreateShareLink] bcrypt: %v", err)
			http.Error(w, "failed to hash password", http.StatusInternalServerError)
			return
		}
	}

	sharelink := model.ShareLink{
		Token:        generateShareToken(),
		ShareTitle:   req.Title,
		CreatedBy:    nowUser,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Duration(req.ExpiresHours) * time.Hour),
		PasswordHash: string(hash),
	}

	tx, err := db.DB.Begin()
	if err != nil {
		log.Printf("[CreateShareLink] begin tx: %v", err)
		http.Error(w, "failed to start tx", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("INSERT INTO share_links (Token, Title, CreatedAt, CreatedBy, ExpiresAt, PasswordHash) VALUES (?, ?, ?, ?, ?, ?)", sharelink.Token, sharelink.ShareTitle, sharelink.CreatedAt.Format(time.RFC3339), sharelink.CreatedBy, sharelink.ExpiresAt.Format(time.RFC3339), sharelink.PasswordHash)
	if err != nil {
		log.Printf("[CreateShareLink] share_links INSERT: %v", err)
		http.Error(w, "failed to create share link", http.StatusInternalServerError)
		return
	}

	for _, id := range req.FileIDs {
		var p int

		err = tx.QueryRow("SELECT Permission FROM file_permissions WHERE UserID = ? AND FileID = ?", nowUser, id).Scan(&p)
		if err == sql.ErrNoRows {
			log.Printf("[CreateShareLink] permission denied user=%s file=%s", nowUser, id)
			continue
		}
		if (p & model.PermDownload) == 0 {
			log.Printf("[CreateShareLink] permission denied user=%s file=%s", nowUser, id)
			continue
		}

		_, err = tx.Exec("INSERT INTO share_link_files (Token, FileID) VALUES (?, ?)", sharelink.Token, id)
		if err != nil {
			log.Printf("[CreateShareLink] share_link_files INSERT [%s]: %v", id, err)
			http.Error(w, "failed to create share link files", http.StatusInternalServerError)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("[CreateShareLink] commit tx: %v", err)
		http.Error(w, "failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		Token string `json:"token"`
	}{sharelink.Token})
}

func getLinkFiles(token string, w http.ResponseWriter) (files []model.File, err error) {
	rows, err := db.DB.Query("SELECT f.ID, f.OriginalName, f.Size, f.MimeType, f.UploadedAt  FROM files f JOIN share_link_files s ON f.ID = s.FileID WHERE s.Token = ?", token)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var f model.File
		err = rows.Scan(&f.ID, &f.OriginalName, &f.Size, &f.MimeType, &f.UploadedAt)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		files = append(files, f)
	}

	return files, err
}

func ShareInfo(w http.ResponseWriter, r *http.Request) {

	var sharelink model.ShareLink

	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "failed to find token param", http.StatusBadRequest)
		return
	}

	var expiresAtStr string
	err := db.DB.QueryRow("SELECT ExpiresAt, PasswordHash, CreatedBy, Title FROM share_links WHERE Token = ?", token).Scan(&expiresAtStr, &sharelink.PasswordHash, &sharelink.CreatedBy, &sharelink.ShareTitle)

	if err == sql.ErrNoRows {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if err != nil {
		log.Printf("[ShareInfo] share_links query [%s]: %v", token, err)
		http.Error(w, "failed to query share link", http.StatusInternalServerError)
		return
	}

	sharelink.ExpiresAt, _ = time.Parse(time.RFC3339, expiresAtStr)
	if time.Now().After(sharelink.ExpiresAt) {
		http.Error(w, "share link expired", http.StatusGone)
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
			return
		} else {

			files, err := getLinkFiles(token, w)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			var resp struct {
				Title     string       `json:"title"`
				ExpiresAt string       `json:"expires_at"`
				CreatedBy string       `json:"created_by"`
				Files     []model.File `json:"files"`
			}
			var un string
			err = db.DB.QueryRow("SELECT Username FROM users WHERE ID = ?", sharelink.CreatedBy).Scan(&un)
			resp.Title = sharelink.ShareTitle
			resp.ExpiresAt = expiresAtStr
			resp.CreatedBy = un
			resp.Files = files

			data, err := json.Marshal(&resp)
			if err != nil {
				log.Printf("[ShareInfo] json marshal: %v", err)
				http.Error(w, "failed to encode response", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(data)
			return
		}
	case http.MethodPost:
		var req struct {
			Password string `json:"password"`
		}
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "failed to decode request", http.StatusBadRequest)
			return
		}

		if req.Password == "" {
			http.Error(w, "failed to find password in request", http.StatusBadRequest)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(sharelink.PasswordHash), []byte(req.Password))
		if err != nil {
			http.Error(w, "failed to authenticate", http.StatusUnauthorized)
			return
		}

		files, err := getLinkFiles(token, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		session := model.ShareSession{
			ShareLinkToken: token,
			Token:          generateID(),
			ExpiresAt:      time.Now().Add(2 * time.Hour),
		}

		_, err = db.DB.Exec("INSERT INTO share_sessions (ShareLinkToken, Token, ExpiresAt) VALUES (?, ?, ?) ", session.ShareLinkToken, session.Token, session.ExpiresAt.Format(time.RFC3339))
		if err != nil {
			log.Printf("[ShareInfo] share_sessions INSERT: %v", err)
			http.Error(w, "failed to create share session", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "share_session",
			Value:    session.Token,
			Expires:  session.ExpiresAt,
			HttpOnly: true,
			Secure:   false,
			Path:     "/",
		})

		var un string
		db.DB.QueryRow("SELECT Username FROM users WHERE ID = ?", sharelink.CreatedBy).Scan(&un)

		var resp struct {
			Title     string       `json:"share_title"`
			ExpiresAt string       `json:"expires_at"`
			CreatedBy string       `json:"created_by"`
			Files     []model.File `json:"files"`
		}
		resp.Title = sharelink.ShareTitle
		resp.ExpiresAt = expiresAtStr
		resp.CreatedBy = un
		resp.Files = files

		data, err := json.Marshal(&resp)
		if err != nil {
			log.Printf("[ShareInfo] json marshal: %v", err)
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func DownloadShareFiles(w http.ResponseWriter, r *http.Request) {

	var sharelink model.ShareLink
	var shareSession model.ShareSession
	var file model.File

	filetoken := r.URL.Query().Get("token")
	if filetoken == "" {
		http.Error(w, "failed to find token param", http.StatusBadRequest)
		return
	}

	var expiresAt_sharelink string
	var hash string
	err := db.DB.QueryRow("SELECT ExpiresAt, PasswordHash FROM share_links WHERE Token = ?", filetoken).Scan(&expiresAt_sharelink, &hash)

	if err == sql.ErrNoRows {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if err != nil {
		log.Printf("[DownloadShareFiles] share_links query [%s]: %v", filetoken, err)
		http.Error(w, "failed to query share link", http.StatusInternalServerError)
		return
	}

	sharelink.ExpiresAt, _ = time.Parse(time.RFC3339, expiresAt_sharelink)

	if time.Now().After(sharelink.ExpiresAt) {
		http.Error(w, "share link expired", http.StatusGone)
		return
	}

	if hash != "" {
		cookie, err := r.Cookie("share_session")
		if err != nil {
			http.Error(w, "failed to authenticate", http.StatusUnauthorized)
			return
		}
		token := cookie.Value

		var expiresAt_sharesession string
		err = db.DB.QueryRow("SELECT ExpiresAt FROM share_sessions WHERE Token = ? AND ShareLinkToken = ?", token, filetoken).Scan(&expiresAt_sharesession)
		if err != nil {
			http.Error(w, "failed to authenticate", http.StatusUnauthorized)
			return
		}

		shareSession.ExpiresAt, _ = time.Parse(time.RFC3339, expiresAt_sharesession)
		if time.Now().After(shareSession.ExpiresAt) {
			_, err = db.DB.Exec("DELETE FROM share_sessions WHERE Token = ?", token)
			if err != nil {
				log.Printf("[DownloadShareFiles] share_sessions DELETE [%s]: %v", token, err)
				http.Error(w, "failed to delete share session", http.StatusInternalServerError)
				return
			}
			http.SetCookie(w, &http.Cookie{
				Name:    "share_session",
				Value:   "",
				Expires: time.Unix(0, 0),
				Path:    "/",
			})
			http.Error(w, "failed to authenticate", http.StatusUnauthorized)
			return
		}
	}

	getFileIDs := r.URL.Query()["ids"]
	if len(getFileIDs) == 0 {
		http.Error(w, "failed to find file ids in request", http.StatusBadRequest)
		return
	}

	rows, err := db.DB.Query("SELECT FileID FROM share_link_files WHERE Token = ?", filetoken)
	if err != nil {
		log.Printf("[DownloadShareFiles] share_link_files query [%s]: %v", filetoken, err)
		http.Error(w, "failed to query share link files", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	vaildIDs := map[string]bool{}

	for rows.Next() {
		var id string
		err = rows.Scan(&id)
		if err != nil {
			log.Printf("[DownloadShareFiles] share_link_files scan: %v", err)
			http.Error(w, "failed to read share link files", http.StatusInternalServerError)
			return
		}
		vaildIDs[id] = true
	}

	filtered := getFileIDs[:0]
	for _, id := range getFileIDs {
		if vaildIDs[id] {
			filtered = append(filtered, id)
		}
	}

	getFileIDs = filtered

	if len(getFileIDs) == 0 {
		http.Error(w, "no valid file", http.StatusBadRequest)
		return
	}

	if len(getFileIDs) == 1 {
		file.ID = getFileIDs[0]
		err = db.DB.QueryRow("SELECT OwnerID, OriginalName, StoredName, Size, MimeType FROM files WHERE ID = ?", file.ID).Scan(&file.OwnerID, &file.OriginalName, &file.StoredName, &file.Size, &file.MimeType)
		if err != nil {
			http.Error(w, "failed to find file", http.StatusNotFound)
			return
		}

		getFile, err := os.Open(filepath.Join(uploadFilesPath, file.StoredName))
		if err != nil {
			log.Printf("[DownloadShareFiles] file open [%s]: %v", file.StoredName, err)
			http.Error(w, "failed to open file", http.StatusInternalServerError)
			return
		}
		defer getFile.Close()

		w.Header().Set("Content-Type", file.MimeType)
		w.Header().Set("Content-Disposition", "attachment; filename=\""+file.OriginalName+"\"")

		_, err = io.Copy(w, getFile)
		if err != nil {
			log.Printf("[DownloadShareFiles] file send [%s]: %v", file.StoredName, err)
			http.Error(w, "failed to send file", http.StatusInternalServerError)
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
				http.Error(w, "failed to find file", http.StatusNotFound)
				return
			}

			getFile, err := os.Open(filepath.Join(uploadFilesPath, file.StoredName))
			if err != nil {
				log.Printf("[DownloadShareFiles] file open [%s]: %v", file.StoredName, err)
				http.Error(w, "failed to open file", http.StatusInternalServerError)
				return
			}

			entry, err := zipWriter.Create(file.OriginalName)
			if err != nil {
				log.Printf("[DownloadShareFiles] zip entry create [%s]: %v", file.OriginalName, err)
				getFile.Close()
				zipWriter.Close()
				return
			}

			_, err = io.Copy(entry, getFile)
			if err != nil {
				log.Printf("[DownloadShareFiles] file send [%s]: %v", file.StoredName, err)
				http.Error(w, "failed to send file", http.StatusInternalServerError)
				getFile.Close()
				return
			}
			getFile.Close()
		}
		zipWriter.Close()
	}
}
