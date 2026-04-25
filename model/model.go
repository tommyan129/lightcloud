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
	ID           string    `json:"id"`
	OwnerID      string    `json:"owner_id"` //User.ID 참조
	OriginalName string    `json:"original_name"`
	StoredName   string    `json:"stored_name"`
	Size         int64     `json:"size"`
	MimeType     string    `json:"mime_type"`
	CreatedAt    time.Time `json:"created_at"`
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
