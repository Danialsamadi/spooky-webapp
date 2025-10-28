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

type ProfileImage struct {
	ID           int       `json:"id"`