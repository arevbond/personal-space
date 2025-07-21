package posts

import (
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

func (s *StorageSuite) TestFind() {
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

func (s *StorageSuite) TestFindBySlug() {
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
