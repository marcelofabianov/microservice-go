package port

import (
	"context"

	"github.com/marcelofabianov/course/internal/user/domain"
)

type RegisterUserInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type RegisterUserOutput struct {
	User *domain.User
}

type RegisterUserUseCase interface {
	Execute(ctx context.Context, input *RegisterUserInput) (*RegisterUserOutput, error)
}
