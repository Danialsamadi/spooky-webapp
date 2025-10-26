package models

import "time"

type Post struct {
	ID        int
	Title     string
	Content   string
	AuthorID  int
	Author    string
	Username  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
