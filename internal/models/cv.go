package models

import "time"

type CV struct {
	ID            int
	Name          string
	Content       []byte
	FileExtension string
	UpdatedAt     time.Time
}
