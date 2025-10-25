package models

import "time"

type User struct {
	Id        int
	Username  string
	Password  string
	Email     string
	CreatedAt time.Time
}
