package model

import "time"

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type File struct {
	ID           string `json:"id"`
	OwnerID      string `json:"owner_id"`  //User.ID 참조
	FolderId     string `json:"folder_id"` //Folder.ID 참조
	OriginalName string `json:"original_name"`
	StoredName   string `json:"stored_name"`
	Size         int64  `json:"size"`
	MimeType     string `json:"mime_type"`
	UploadedAt   string `json:"uploaded_at"`
	OwnerName    string `json:"owner_name,omitempty"` // shared 파일에만 채워짐
}

type Folder struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`
	ParentID  string    `json:"parent_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type FileListResponse struct {
	Mine   []File `json:"mine"`
	Shared []File `json:"shared"`
}

const (
	PermRead = 1 << iota
	PermDownload
	PermWrite
	PermDelete
	PermManage
)

type FilePermission struct {
	FileID     string `json:"file_id"`
	UserID     string `json:"user_id"`
	Permission int    `json:"permission"`
}

type ShareLink struct {
	Token        string    `json:"token"`
	ShareTitle   string    `json:"share_title"`
	CreatedBy    string    `json:"created_by"` //User.ID 참조
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	PasswordHash string    `json:"-"`
}

type ShareLinkFile struct {
	Token  string `json:"token"`
	FileID string `json:"file_id"`
}

type Session struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"` //User.ID 참조
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type ShareSession struct {
	ShareLinkToken string    `json:"share_link_token"`
	Token          string    `json:"token"`
	ExpiresAt      time.Time `json:"expires_at"`
}
