package server

import "github.com/arevbond/arevbond-blog/internal/service/blog/domain"

type PostsPageData struct {
	Posts        []*domain.Post
	IsAdmin      bool
	HasNextPages bool
	NextOffset   int
}
