package posts

import (
	"fmt"
	"github.com/arevbond/arevbond-blog/internal/service/blog/storage"
)

func (s *StorageSuite) TestCategoriesAll() {
	repo := storage.NewCategoriesRepo(s.log, s.conn)

	result, err := repo.All(s.ctx)
	s.Require().NoError(err)

	defaultCategories := map[string]bool{
		"Книги":      false,
		"Технологии": false,
		"Разное":     false,
	}

	for _, category := range result {
		defaultCategories[category.Name] = true
	}

	for name, ok := range defaultCategories {
		if !ok {
			s.Assert().Fail(fmt.Sprintf("category '%s' doesn't exists", name))
		}
	}
}
