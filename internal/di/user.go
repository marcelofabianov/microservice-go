package di

import (
	"go.uber.org/fx"

	"github.com/marcelofabianov/course/internal/user/handler"
	"github.com/marcelofabianov/course/internal/user/port"
	"github.com/marcelofabianov/course/internal/user/storage"
	"github.com/marcelofabianov/course/internal/user/usecase"
	"github.com/marcelofabianov/course/pkg/crypto"
)

var UserModule = fx.Module("user",
	fx.Provide(
		crypto.NewArgon2Hasher,
		func(h *crypto.Argon2Hasher) port.PasswordHasherPort { return h },
		storage.NewPostgresUserRepository,
		func(r *storage.PostgresUserRepository) port.CreateUserRepositoryPort { return r },
		fx.Annotate(
			usecase.NewRegisterUserUseCase,
			fx.As(new(port.RegisterUserUseCase)),
		),
		handler.NewRegisterUserHandler,
		AsRouter(handler.NewUserRouter),
	),
)
