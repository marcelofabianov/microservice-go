package port

import (
	"context"

	"github.com/marcelofabianov/course/internal/user/domain"
)

type CreateUserRepositoryPort interface {
	CreateUser(ctx context.Context, user *domain.User) error
}

type UserRepositoryPort interface {
	CreateUserRepositoryPort
}
