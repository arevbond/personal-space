package domain

import (
	"time"
)

type Post struct {
	ID          int       `db:"id"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	Content     []byte    `db:"content"`
	Extension   string    `db:"extension"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type PostParams struct {
	Title       string
	Description string
	Filename    string
	Content     []byte
}
