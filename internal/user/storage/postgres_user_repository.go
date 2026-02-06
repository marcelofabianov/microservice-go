package storage

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/marcelofabianov/course/internal/user/domain"
	"github.com/marcelofabianov/course/pkg/database"
)

const defaultExecTimeout = 5 * time.Second

const createUserQuery = `
	INSERT INTO users (
		id, name, email, phone, hashed_password, role, is_active,
		audit_created_at, audit_created_by,
		audit_updated_at, audit_updated_by,
		audit_archived_at, audit_deleted_at, audit_version
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
`

type PostgresUserRepository struct {
	db *database.DB
}

func NewPostgresUserRepository(db *database.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	execCtx, cancel := context.WithTimeout(ctx, defaultExecTimeout)
	defer cancel()

	_, err := r.db.DB().ExecContext(execCtx, createUserQuery,
		user.ID,
		user.Name,
		user.Email,
		user.Phone,
		user.PasswordHash,
		user.Role,
		user.IsActive,
		user.Audit.CreatedAt,
		user.Audit.CreatedBy,
		user.Audit.UpdatedAt,
		user.Audit.UpdatedBy,
		user.Audit.ArchivedAt,
		user.Audit.DeletedAt,
		user.Audit.Version,
	)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		if strings.Contains(pgErr.ConstraintName, "email") {
			return domain.NewErrUserEmailAlreadyExists()
		}
		if strings.Contains(pgErr.ConstraintName, "phone") {
			return domain.NewErrUserPhoneAlreadyExists()
		}
	}

	return domain.NewErrUserFailedCreateUser()
}
