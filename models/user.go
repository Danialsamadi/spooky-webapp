package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Password     string    `json:"password"`
	Email        string    `json:"email"`
	Bio          string    `json:"bio"`
	ProfileImage string    `json:"profile_image"`
	Location     string    `json:"location"`
	Website      string    `json:"website"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ProfileImage struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	Filename     string    `json:"filename"`
	OriginalName string    `json:"original_name"`
	FilePath     string    `json:"file_path"`
	FileSize     int       `json:"file_size"`
	MimeType     string    `json:"mime_type"`
	CreatedAt    time.Time `json:"created_at"`
}
