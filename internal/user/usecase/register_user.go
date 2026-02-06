package usecase

import (
	"context"

	"github.com/marcelofabianov/wisp"

	"github.com/marcelofabianov/course/internal/user/domain"
	"github.com/marcelofabianov/course/internal/user/port"
)

type RegisterUserUseCase struct {
	repo   port.CreateUserRepositoryPort
	hasher port.PasswordHasherPort
}

func NewRegisterUserUseCase(repo port.CreateUserRepositoryPort, hasher port.PasswordHasherPort) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		repo:   repo,
		hasher: hasher,
	}
}

func (uc *RegisterUserUseCase) Execute(ctx context.Context, input *port.RegisterUserInput) (*port.RegisterUserOutput, error) {
	passwordHash, err := uc.hasher.Hash(input.Password)
	if err != nil {
		return nil, domain.NewErrUserFailedHashPassword()
	}

	hash, err := wisp.NewNonEmptyString(passwordHash)
	if err != nil {
		return nil, domain.NewErrUserFailedHashPassword()
	}

	user, err := domain.NewUser(&domain.NewUserInput{
		Name:  input.Name,
		Email: input.Email,
		Role:  input.Role,
		Phone: input.Phone,
	}, hash, wisp.AuditUser("system"))
	if err != nil {
		return nil, err
	}

	if err := uc.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return &port.RegisterUserOutput{
		User: user,
		Response: &port.RegisterUserResponse{
			ID:        user.ID.String(),
			Name:      user.Name.String(),
			Email:     user.Email.String(),
			Phone:     user.Phone.String(),
			Role:      user.Role.String(),
			IsActive:  user.IsActive,
			CreatedAt: user.Audit.CreatedAt.Time().Format("2006-01-02T15:04:05Z07:00"),
			Version:   user.Audit.Version.Int(),
		},
	}, nil
}
