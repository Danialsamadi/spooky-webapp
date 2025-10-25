package models

import "time"

type Post struct {
	Id        int
	Title     string
	Content   string
	AuthorId  int
	Author    string
	Username  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
