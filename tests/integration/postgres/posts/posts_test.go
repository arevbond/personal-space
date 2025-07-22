package posts

import (
	"fmt"
	"github.com/arevbond/arevbond-blog/internal/service/blog/domain"
	"github.com/arevbond/arevbond-blog/internal/service/blog/storage"
	"github.com/arevbond/arevbond-blog/internal/service/errs"
	"time"
)

// insertTestPost - helper function, which create single post in database
func (s *StorageSuite) insertTestPost(slug string) (*domain.Post, error) {
	post := &domain.Post{
		ID:           0,
		Title:        "title",
		Description:  "description",
		Content:      []byte("100001"),
		Extension:    ".md",
		Slug:         slug,
		CategoryID:   1,
		CategoryName: "Книги",
		IsPublished:  true,
		CreatedAt:    time.Date(2005, 1, 1, 1, 0, 0, 0, time.Local),
		UpdatedAt:    time.Date(2005, 1, 1, 1, 0, 0, 0, time.Local),
	}

	query := `
		INSERT INTO posts (title, description, content, extension, slug, is_published, 
		                   category_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id;`

	args := []any{post.Title, post.Description, post.Content, post.Extension, post.Slug,
		post.IsPublished, post.CategoryID, post.CreatedAt, post.UpdatedAt}

	row := s.conn.QueryRowContext(s.ctx, query, args...)
	if err := row.Scan(&post.ID); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *StorageSuite) TestPostsCreate() {
	repo := storage.NewPostsRepo(s.log, s.conn)

	tests := []struct {
		name          string
		post          *domain.Post
		expectedError error
	}{
		{
			name: "successfully creates a post",
			post: &domain.Post{
				ID:          0,
				Title:       "title",
				Description: "desc",
				Content:     []byte("1110"),
				Slug:        "slug",
				CategoryID:  1,
				Extension:   ".md",
				CreatedAt:   time.Date(2005, 1, 1, 1, 0, 0, 0, time.Local),
				UpdatedAt:   time.Date(2005, 1, 1, 1, 0, 0, 0, time.Local),
			},
		},
		{
			name: "duplicate error",
			post: &domain.Post{
				ID:          0,
				Title:       "title2",
				Description: "desc2",
				Content:     []byte("11101"),
				Slug:        "slug",
				CategoryID:  1,
				Extension:   ".md",
			},
			expectedError: errs.ErrDuplicate,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := repo.Create(s.ctx, tt.post)
			if tt.expectedError == nil {
				s.Assert().NoError(err)
				s.Assert().NotEqual(0, tt.post.ID)

				query := `SELECT p.id, title, description, content, extension, slug, is_published, category_id,
							   c.name as category_name, created_at, updated_at
						FROM posts p
						INNER JOIN categories c ON p.category_id = c.id
						WHERE p.id = $1; `

				var expectedPost domain.Post

				s.Require().NoError(s.conn.GetContext(s.ctx, &expectedPost, query, tt.post.ID), "should fetch post")

				s.Assert().Equal(expectedPost.Title, tt.post.Title, "title should be equal")
				s.Assert().Equal(expectedPost.Description, tt.post.Description, "description should be equal")
				s.Assert().ElementsMatch(expectedPost.Content, tt.post.Content, "content should be equal")
				s.Assert().Equal(expectedPost.Slug, tt.post.Slug, "slug should be equal")
				s.Assert().Equal(expectedPost.CategoryID, tt.post.CategoryID, "category id should be equal")
				s.Assert().Equal(expectedPost.Extension, tt.post.Extension, "extension should be equal")
				s.Assert().Equal(expectedPost.CreatedAt, tt.post.CreatedAt, "created time should be equal")
				s.Assert().Equal(expectedPost.UpdatedAt, tt.post.UpdatedAt, "updated time should be equal")
			} else {
				s.Require().Error(err)
				s.Assert().ErrorIs(err, tt.expectedError)
			}
		})
	}
}

func (s *StorageSuite) TestPostsUpdate() {
	post, err := s.insertTestPost("slug")
	s.Require().NoError(err)

	params := domain.UpdatePostParams{
		ID:          post.ID,
		Title:       post.Title + "123",
		Slug:        post.Slug + "123",
		Description: post.Description + "123",
		CategoryID:  2,
		Content:     []byte("123123"),
	}

	repo := storage.NewPostsRepo(s.log, s.conn)

	err = repo.Update(s.ctx, params)
	s.Require().NoError(err)

	query := `SELECT p.id, title, description, content, extension, slug, is_published, category_id,
							   c.name as category_name, created_at, updated_at
						FROM posts p
						INNER JOIN categories c ON p.category_id = c.id
						WHERE p.id = $1; `

	var postInDB domain.Post

	s.Require().NoError(s.conn.GetContext(s.ctx, &postInDB, query, post.ID), "should fetch post")

	s.Assert().Equal(params.Title, postInDB.Title)
	s.Assert().Equal(params.Slug, postInDB.Slug)
	s.Assert().Equal(params.Description, postInDB.Description)
	s.Assert().Equal(params.CategoryID, postInDB.CategoryID)
	s.Assert().Equal(params.Content, postInDB.Content)
}

func (s *StorageSuite) TestPostsFind() {
	expectedPost, err := s.insertTestPost("slug")
	s.Require().NoError(err)

	repo := storage.NewPostsRepo(s.log, s.conn)

	s.Run("successfully find", func() {
		resultPost, err2 := repo.Find(s.ctx, expectedPost.ID)
		s.Require().NoError(err2)

		s.Assert().Equal(expectedPost.Title, resultPost.Title)
		s.Assert().Equal(expectedPost.Description, resultPost.Description)
		s.Assert().Equal(expectedPost.Content, resultPost.Content)
		s.Assert().Equal(expectedPost.Slug, resultPost.Slug, "slug should be equal")
		s.Assert().Equal(expectedPost.CategoryID, resultPost.CategoryID, "category id should be equal")
		s.Assert().Equal(expectedPost.Extension, resultPost.Extension, "extension should be equal")
		s.Assert().Equal(expectedPost.CreatedAt, resultPost.CreatedAt, "created time should be equal")
		s.Assert().Equal(expectedPost.UpdatedAt, resultPost.UpdatedAt, "updated time should be equal")
	})

	s.Run("empty result", func() {
		resultPost, err2 := repo.Find(s.ctx, 1234)
		s.Require().Error(err2)
		s.Nil(resultPost)
	})
}

func (s *StorageSuite) TestPostsFindBySlug() {
	expectedPost, err := s.insertTestPost("slug")
	s.Require().NoError(err)

	repo := storage.NewPostsRepo(s.log, s.conn)

	s.Run("successfully find", func() {
		resultPost, err2 := repo.FindBySlug(s.ctx, "slug")
		s.Require().NoError(err2)

		s.Assert().Equal(expectedPost.Title, resultPost.Title)
		s.Assert().Equal(expectedPost.Description, resultPost.Description)
		s.Assert().Equal(expectedPost.Content, resultPost.Content)
		s.Assert().Equal(expectedPost.Slug, resultPost.Slug, "slug should be equal")
		s.Assert().Equal(expectedPost.CategoryID, resultPost.CategoryID, "category id should be equal")
		s.Assert().Equal(expectedPost.Extension, resultPost.Extension, "extension should be equal")
		s.Assert().Equal(expectedPost.CreatedAt, resultPost.CreatedAt, "created time should be equal")
		s.Assert().Equal(expectedPost.UpdatedAt, resultPost.UpdatedAt, "updated time should be equal")
	})

	s.Run("empty result", func() {
		resultPost, err2 := repo.FindBySlug(s.ctx, "invalid")
		s.Require().Error(err2)
		s.Nil(resultPost)
	})
}

func (s *StorageSuite) TestPostsAll() {
	repo := storage.NewPostsRepo(s.log, s.conn)

	// isPublished (true) - 3
	// isPublished (false) -2
	posts := make([]*domain.Post, 5)

	for i := 0; i < 5; i++ {
		posts[i] = &domain.Post{
			Title:       fmt.Sprintf("Post %d", i+1),
			Description: fmt.Sprintf("Description %d", i+1),
			Content:     []byte(fmt.Sprintf("Content %d", i+1)),
			Slug:        fmt.Sprintf("post-%d", i+1),
			CategoryID:  1,
			Extension:   ".md",
			IsPublished: i%2 == 0,
			CreatedAt:   time.Now().Add(-time.Duration(i) * time.Hour),
			UpdatedAt:   time.Now(),
		}
		s.Require().NoError(repo.Create(s.ctx, posts[i]))
	}

	s.Run("posts only published", func() {
		result, err := repo.All(s.ctx, 10, 0, true)
		s.Require().NoError(err)

		s.Require().Equal(3, len(result), "should return all published posts")
	})

	s.Run("all posts", func() {
		result, err := repo.All(s.ctx, 10, 0, false)
		s.Require().NoError(err)

		s.Require().Equal(len(result), len(result), "should return all published posts")
	})
}

func (s *StorageSuite) TestPostsAll_Pagination() {
	repo := storage.NewPostsRepo(s.log, s.conn)

	posts := make([]*domain.Post, 5)

	for i := 0; i < 5; i++ {
		posts[i] = &domain.Post{
			Title:       fmt.Sprintf("Post %d", i+1),
			Description: fmt.Sprintf("Description %d", i+1),
			Content:     []byte(fmt.Sprintf("Content %d", i+1)),
			Slug:        fmt.Sprintf("post-%d", i+1),
			CategoryID:  1,
			Extension:   ".md",
			IsPublished: true,
			CreatedAt:   time.Now().Add(-time.Duration(i) * time.Hour),
			UpdatedAt:   time.Now(),
		}
		s.Require().NoError(repo.Create(s.ctx, posts[i]))
	}

	tests := []struct {
		name          string
		limit         int
		offset        int
		expectedCount int
		expectedFirst string
	}{
		{
			name:          "first page - limit 2",
			limit:         2,
			offset:        0,
			expectedCount: 2,
			expectedFirst: "Post 1",
		},
		{
			name:          "second page - limit 2, offset 2",
			limit:         2,
			offset:        2,
			expectedCount: 2,
			expectedFirst: "Post 3",
		},
		{
			name:          "last page - limit 2, offset 4",
			limit:         2,
			offset:        4,
			expectedCount: 1,
			expectedFirst: "Post 5",
		},
		{
			name:          "offset beyond data",
			limit:         2,
			offset:        10,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result, err := repo.All(s.ctx, tt.limit, tt.offset, true)
			s.Require().NoError(err)
			s.Assert().Len(result, tt.expectedCount)

			if tt.expectedCount > 0 {
				s.Assert().Equal(tt.expectedFirst, result[0].Title)
			}
		})
	}
}

func (s *StorageSuite) TestPostsSetPublicationStatus() {
	post, err := s.insertTestPost("slug")
	s.Require().NoError(err)

	repo := storage.NewPostsRepo(s.log, s.conn)

	err = repo.SetPublicationStatus(s.ctx, post.ID, !post.IsPublished)
	s.Require().NoError(err)

	postInDB, err := repo.FindBySlug(s.ctx, "slug")
	s.Require().NoError(err)

	s.Assert().Equal(!post.IsPublished, postInDB.IsPublished)
}

func (s *StorageSuite) TestPostsDelete() {
	post, err := s.insertTestPost("slug")
	s.Require().NoError(err)

	_, err = s.insertTestPost("slug2")
	s.Require().NoError(err)

	repo := storage.NewPostsRepo(s.log, s.conn)

	err = repo.Delete(s.ctx, post.ID)
	s.Require().NoError(err)

	var count int
	query := `SELECT count(title) FROM POSTS `

	s.Require().NoError(s.conn.GetContext(s.ctx, &count, query))

	s.Assert().Equal(1, count)
}

func (s *StorageSuite) TestPostsAllWithCategory() {
	repo := storage.NewPostsRepo(s.log, s.conn)

	// Create test data:
	// Category 1: 2 posts (1 published, 1 unpublished)
	// Category 2: 3 posts (2 published, 1 unpublished)
	posts := make([]*domain.Post, 5)

	for i := 0; i < 5; i++ {
		categoryID := 1
		if i >= 2 { // First 2 posts go to category 1, rest to category 2
			categoryID = 2
		}

		posts[i] = &domain.Post{
			Title:       fmt.Sprintf("Post %d", i+1),
			Description: fmt.Sprintf("Description %d", i+1),
			Content:     []byte(fmt.Sprintf("Content %d", i+1)),
			Slug:        fmt.Sprintf("post-%d", i+1),
			CategoryID:  categoryID,
			Extension:   ".md",
			IsPublished: i != 1 && i != 4, // Post 2 and Post 5 are unpublished
			CreatedAt:   time.Now().Add(-time.Duration(i) * time.Hour),
			UpdatedAt:   time.Now(),
		}
		s.Require().NoError(repo.Create(s.ctx, posts[i]))
	}

	s.Run("published posts only from category 1", func() {
		result, err := repo.AllWithCategory(s.ctx, 10, 0, true, 1)
		s.Require().NoError(err)
		s.Assert().Len(result, 1, "should return 1 published post from category 1")
		s.Assert().Equal(1, result[0].CategoryID, "all posts should be from category 1")
		s.Assert().True(result[0].IsPublished, "all posts should be published")
	})

	s.Run("published posts only from category 2", func() {
		result, err := repo.AllWithCategory(s.ctx, 10, 0, true, 2)
		s.Require().NoError(err)
		s.Assert().Len(result, 2, "should return 2 published posts from category 2")
		for _, post := range result {
			s.Assert().Equal(2, post.CategoryID, "all posts should be from category 2")
			s.Assert().True(post.IsPublished, "all posts should be published")
		}
	})

	s.Run("all posts from category 1", func() {
		result, err := repo.AllWithCategory(s.ctx, 10, 0, false, 1)
		s.Require().NoError(err)
		s.Assert().Len(result, 2, "should return all posts from category 1")
		for _, post := range result {
			s.Assert().Equal(1, post.CategoryID, "all posts should be from category 1")
		}
	})

	s.Run("all posts from category 2", func() {
		result, err := repo.AllWithCategory(s.ctx, 10, 0, false, 2)
		s.Require().NoError(err)
		s.Assert().Len(result, 3, "should return all posts from category 2")
		for _, post := range result {
			s.Assert().Equal(2, post.CategoryID, "all posts should be from category 2")
		}
	})
}

func (s *StorageSuite) TestPostsAllWithCategory_Pagination() {
	repo := storage.NewPostsRepo(s.log, s.conn)

	// Create 5 posts in category 1 for pagination testing
	posts := make([]*domain.Post, 5)
	for i := 0; i < 5; i++ {
		posts[i] = &domain.Post{
			Title:       fmt.Sprintf("Category1 Post %d", i+1),
			Description: fmt.Sprintf("Description %d", i+1),
			Content:     []byte(fmt.Sprintf("Content %d", i+1)),
			Slug:        fmt.Sprintf("category1-post-%d", i+1),
			CategoryID:  1, // All in category 1
			Extension:   ".md",
			IsPublished: true,
			CreatedAt:   time.Now().Add(-time.Duration(i) * time.Hour),
			UpdatedAt:   time.Now(),
		}
		s.Require().NoError(repo.Create(s.ctx, posts[i]))
	}

	tests := []struct {
		name          string
		limit         int
		offset        int
		categoryID    int
		expectedCount int
		expectedFirst string
	}{
		{
			name:          "first page - limit 2, category 1",
			limit:         2,
			offset:        0,
			categoryID:    1,
			expectedCount: 2,
			expectedFirst: "Category1 Post 1", // Most recent
		},
		{
			name:          "second page - limit 2, offset 2, category 1",
			limit:         2,
			offset:        2,
			categoryID:    1,
			expectedCount: 2,
			expectedFirst: "Category1 Post 3",
		},
		{
			name:          "last page - limit 2, offset 4, category 1",
			limit:         2,
			offset:        4,
			categoryID:    1,
			expectedCount: 1,
			expectedFirst: "Category1 Post 5",
		},
		{
			name:          "offset beyond data",
			limit:         2,
			offset:        10,
			categoryID:    1,
			expectedCount: 0,
		},
		{
			name:          "non-existent category",
			limit:         10,
			offset:        0,
			categoryID:    999,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result, err := repo.AllWithCategory(s.ctx, tt.limit, tt.offset, true, tt.categoryID)
			s.Require().NoError(err)
			s.Assert().Len(result, tt.expectedCount)

			if tt.expectedCount > 0 {
				s.Assert().Equal(tt.expectedFirst, result[0].Title)
				for _, post := range result {
					s.Assert().Equal(tt.categoryID, post.CategoryID)
				}
			}
		})
	}
}
