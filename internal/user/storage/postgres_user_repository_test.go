package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/marcelofabianov/course/config"
	"github.com/marcelofabianov/course/pkg/database"
)

func TestNewPostgresUserRepository(t *testing.T) {
	t.Run("creates repository with valid database", func(t *testing.T) {
		cfg, err := config.Load()
		require.NoError(t, err)

		db, err := database.New(cfg)
		require.NoError(t, err)

		repo := NewPostgresUserRepository(db)

		assert.NotNil(t, repo)
		assert.NotNil(t, repo.db)
	})
}
