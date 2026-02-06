package port

import (
	"context"

	"github.com/marcelofabianov/course/internal/user/domain"
)

type RegisterUserInput struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email,max=254"`
	Password string `json:"password" validate:"required,min=8,max=72"`
	Phone    string `json:"phone" validate:"required"`
	Role     string `json:"role" validate:"required"`
}

type RegisterUserResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	Version   int    `json:"version"`
}

type RegisterUserOutput struct {
	User     *domain.User
	Response *RegisterUserResponse
}

type RegisterUserUseCase interface {
	Execute(ctx context.Context, input *RegisterUserInput) (*RegisterUserOutput, error)
}
