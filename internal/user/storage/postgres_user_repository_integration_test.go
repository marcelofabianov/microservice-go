//go:build integration

package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/marcelofabianov/wisp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/marcelofabianov/course/config"
	"github.com/marcelofabianov/course/internal/user/domain"
	"github.com/marcelofabianov/course/pkg/database"
)

func setupRepository(t *testing.T) *PostgresUserRepository {
	t.Helper()

	cfg, err := config.Load()
	require.NoError(t, err)

	db, err := database.New(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = db.Connect(ctx)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = db.Close()
	})

	return NewPostgresUserRepository(db)
}

func createTestUser(t *testing.T, email, phone string) *domain.User {
	t.Helper()

	hash, err := wisp.NewNonEmptyString("$argon2id$v=19$m=65536,t=3,p=4$c2FsdA$a2V5")
	require.NoError(t, err)

	user, err := domain.NewUser(&domain.NewUserInput{
		Name:  "Test User",
		Email: email,
		Phone: phone,
		Role:  "common",
	}, hash, wisp.AuditUser("system"))
	require.NoError(t, err)

	return user
}

func cleanupUser(t *testing.T, repo *PostgresUserRepository, userID wisp.UUID) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _ = repo.db.DB().ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)
}

func TestPostgresUserRepository_CreateUser(t *testing.T) {
	t.Run("creates user successfully", func(t *testing.T) {
		repo := setupRepository(t)
		user := createTestUser(t, "integration-test@example.com", "+5511900000001")
		t.Cleanup(func() { cleanupUser(t, repo, user.ID) })

		err := repo.CreateUser(context.Background(), user)

		assert.NoError(t, err)
	})

	t.Run("returns error for duplicate email", func(t *testing.T) {
		repo := setupRepository(t)
		user1 := createTestUser(t, "duplicate-email@example.com", "+5511900000002")
		t.Cleanup(func() { cleanupUser(t, repo, user1.ID) })

		err := repo.CreateUser(context.Background(), user1)
		require.NoError(t, err)

		user2 := createTestUser(t, "duplicate-email@example.com", "+5511900000003")
		t.Cleanup(func() { cleanupUser(t, repo, user2.ID) })

		err = repo.CreateUser(context.Background(), user2)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrUserEmailAlreadyExists))
	})

	t.Run("returns error for duplicate phone", func(t *testing.T) {
		repo := setupRepository(t)
		user1 := createTestUser(t, "phone-test-1@example.com", "+5511900000004")
		t.Cleanup(func() { cleanupUser(t, repo, user1.ID) })

		err := repo.CreateUser(context.Background(), user1)
		require.NoError(t, err)

		user2 := createTestUser(t, "phone-test-2@example.com", "+5511900000004")
		t.Cleanup(func() { cleanupUser(t, repo, user2.ID) })

		err = repo.CreateUser(context.Background(), user2)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrUserPhoneAlreadyExists))
	})

	t.Run("returns error for canceled context", func(t *testing.T) {
		repo := setupRepository(t)
		user := createTestUser(t, "canceled-ctx@example.com", "+5511900000005")

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := repo.CreateUser(ctx, user)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, domain.ErrUserFailedCreateUser))
	})
}
