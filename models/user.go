package models

import "time"

type User struct {
	ID             int       `json:"id"`
	Username       string    `json:"username"`
	Password       string    `json:"password"`
	Email          string    `json:"email"`
	Bio            string    `json:"bio"`
	ProfileImage   string    `json:"profile_image"`
	Location       string    `json:"location"`
	Website        string    `json:"website"`
	InvitationCode string    `json:"invitation_code"`
	InvitedBy      *int      `json:"invited_by"`
	IsAdmin        bool      `json:"is_admin"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type InvitationCode struct {
	ID        int       `json:"id"`
	Code      string    `json:"code"`
	CreatedBy int       `json:"created_by"`
	UsedBy    *int      `json:"used_by"`
	IsUsed    bool      `json:"is_used"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UsedAt    *time.Time `json:"used_at"`
}

type ProfileImage struct {
	ID           int       `json:"id"`