package domain

import (
	"time"
)

type Post struct {
	ID           int       `db:"id"`
	Title        string    `db:"title"`
	Description  string    `db:"description"`
	Content      []byte    `db:"content"`
	Extension    string    `db:"extension"`
	IsPublished  bool      `db:"is_published"`
	Slug         string    `db:"slug"`
	CategoryID   int       `db:"category_id"`
	CategoryName string    `db:"category_name"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type Category struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type SelectPostParams struct {
	Limit      int
	Offset     int
	IsAdmin    bool
	CategoryID int
}

type CreatePostParams struct {
	Title       string
	Slug        string
	Description string
	Filename    string
	CategoryID  int
	IsPublished bool
	Content     []byte
}
