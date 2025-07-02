package server

import "github.com/arevbond/arevbond-blog/internal/service/blog/domain"

type PostsPageData struct {
	Categories []*domain.Category
	PostsData
}

type PostsData struct {
	SelectedCategoryID int
	Posts              []*domain.Post
	IsAdmin            bool
	HasNextPages       bool
	NextOffset         int
}
