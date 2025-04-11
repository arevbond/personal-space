package models

import "time"

type Resume struct {
	ID            int
	Name          string
	Content       []byte
	FileExtension string
	UpdatedAt     time.Time
}
