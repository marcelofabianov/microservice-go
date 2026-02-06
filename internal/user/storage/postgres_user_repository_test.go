package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/marcelofabianov/course/pkg/database"
)

func TestNewPostgresUserRepository(t *testing.T) {
	t.Run("creates repository with valid database", func(t *testing.T) {
		repo := NewPostgresUserRepository(&database.DB{})

		assert.NotNil(t, repo)
	})
}
