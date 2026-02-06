package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/marcelofabianov/course/internal/user/domain"
	"github.com/marcelofabianov/course/internal/user/port"
)

type mockRepository struct {
	err error
}

func (m *mockRepository) CreateUser(_ context.Context, _ *domain.User) error {
	return m.err
}

type mockHasher struct {
	hash string
	err  error
}

func (m *mockHasher) Hash(_ string) (string, error) {
	return m.hash, m.err
}

func validInput() *port.RegisterUserInput {
	return &port.RegisterUserInput{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "Test@123!",
		Phone:    "+5511999999999",
		Role:     "common",
	}
}

func TestRegisterUserUseCase_Execute(t *testing.T) {
	t.Run("creates user successfully", func(t *testing.T) {
		repo := &mockRepository{}
		hasher := &mockHasher{hash: "$argon2id$v=19$m=65536,t=3,p=4$c2FsdA$a2V5"}
		uc := NewRegisterUserUseCase(repo, hasher)

		output, err := uc.Execute(context.Background(), validInput())

		require.NoError(t, err)
		require.NotNil(t, output)
		assert.NotNil(t, output.User)
		assert.NotNil(t, output.Response)
		assert.Equal(t, "John Doe", output.Response.Name)
		assert.Equal(t, "john@example.com", output.Response.Email)
		assert.Equal(t, "5511999999999", output.Response.Phone)
		assert.Equal(t, "common", output.Response.Role)
		assert.True(t, output.Response.IsActive)
		assert.Equal(t, 1, output.Response.Version)
		assert.NotEmpty(t, output.Response.ID)
		assert.NotEmpty(t, output.Response.CreatedAt)
	})

	t.Run("returns error when hash fails", func(t *testing.T) {
		repo := &mockRepository{}
		hasher := &mockHasher{err: errors.New("hash failed")}
		uc := NewRegisterUserUseCase(repo, hasher)

		output, err := uc.Execute(context.Background(), validInput())

		assert.Nil(t, output)
		assert.True(t, errors.Is(err, domain.ErrUserFailedHashPassword))
	})

	t.Run("returns error when domain validation fails", func(t *testing.T) {
		repo := &mockRepository{}
		hasher := &mockHasher{hash: "$argon2id$v=19$m=65536,t=3,p=4$c2FsdA$a2V5"}
		uc := NewRegisterUserUseCase(repo, hasher)

		input := validInput()
		input.Email = "invalid-email"

		output, err := uc.Execute(context.Background(), input)

		assert.Nil(t, output)
		assert.True(t, errors.Is(err, domain.ErrUserInvalidEmail))
	})

	t.Run("returns error when repository fails with duplicate email", func(t *testing.T) {
		repo := &mockRepository{err: domain.NewErrUserEmailAlreadyExists("john@example.com")}
		hasher := &mockHasher{hash: "$argon2id$v=19$m=65536,t=3,p=4$c2FsdA$a2V5"}
		uc := NewRegisterUserUseCase(repo, hasher)

		output, err := uc.Execute(context.Background(), validInput())

		assert.Nil(t, output)
		assert.True(t, errors.Is(err, domain.ErrUserEmailAlreadyExists))
	})

	t.Run("returns error when domain rejects invalid name", func(t *testing.T) {
		repo := &mockRepository{}
		hasher := &mockHasher{hash: "$argon2id$v=19$m=65536,t=3,p=4$c2FsdA$a2V5"}
		uc := NewRegisterUserUseCase(repo, hasher)

		input := validInput()
		input.Name = ""

		output, err := uc.Execute(context.Background(), input)

		assert.Nil(t, output)
		assert.True(t, errors.Is(err, domain.ErrUserInvalidName))
	})
}
