package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/marcelofabianov/fault"
	"github.com/marcelofabianov/wisp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/marcelofabianov/course/internal/user/domain"
	"github.com/marcelofabianov/course/internal/user/port"
	"github.com/marcelofabianov/course/pkg/logger"
	"github.com/marcelofabianov/course/pkg/validation"
)

type mockUseCase struct {
	output *port.RegisterUserOutput
	err    error
}

func (m *mockUseCase) Execute(_ context.Context, _ *port.RegisterUserInput) (*port.RegisterUserOutput, error) {
	return m.output, m.err
}

func newTestValidator(t *testing.T) validation.Validator {
	t.Helper()
	log := logger.New(&logger.Config{
		Level:  logger.LevelError,
		Format: logger.FormatText,
	})
	return validation.New(log, nil)
}

func validRequestBody() []byte {
	body, _ := json.Marshal(map[string]string{
		"name":     "John Doe",
		"email":    "john@example.com",
		"password": "Test@123!",
		"phone":    "+5511999999999",
		"role":     "common",
	})
	return body
}

func TestRegisterUserHandler_Handle(t *testing.T) {
	t.Run("returns 201 on success without password hash", func(t *testing.T) {
		hash, _ := wisp.NewNonEmptyString("$argon2id$v=19$m=65536,t=3,p=4$c2FsdA$a2V5")
		user, err := domain.NewUser(&domain.NewUserInput{
			Name:  "John Doe",
			Email: "john@example.com",
			Phone: "+5511999999999",
			Role:  "common",
		}, hash, wisp.AuditUser("system"))
		require.NoError(t, err)

		uc := &mockUseCase{output: &port.RegisterUserOutput{
			User: user,
			Response: &port.RegisterUserResponse{
				ID:        user.ID.String(),
				Name:      "John Doe",
				Email:     "john@example.com",
				Phone:     "5511999999999",
				Role:      "common",
				IsActive:  true,
				CreatedAt: "2026-02-06T00:00:00Z",
				Version:   1,
			},
		}}
		handler := NewRegisterUserHandler(uc, newTestValidator(t))

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(validRequestBody()))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Handle(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.NotContains(t, w.Body.String(), "PasswordHash")
		assert.NotContains(t, w.Body.String(), "argon2id")
		assert.Contains(t, w.Body.String(), "john@example.com")
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		uc := &mockUseCase{}
		handler := NewRegisterUserHandler(uc, newTestValidator(t))

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Handle(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 for validation error", func(t *testing.T) {
		uc := &mockUseCase{}
		handler := NewRegisterUserHandler(uc, newTestValidator(t))

		body, _ := json.Marshal(map[string]string{
			"name": "",
		})

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Handle(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns error from use case domain error", func(t *testing.T) {
		domainErr := fault.New("domain violation", fault.WithCode(fault.DomainViolation))
		uc := &mockUseCase{err: domainErr}
		handler := NewRegisterUserHandler(uc, newTestValidator(t))

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(validRequestBody()))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Handle(w, req)

		assert.True(t, w.Code >= 400)
	})

	t.Run("returns 409 for duplicate email", func(t *testing.T) {
		uc := &mockUseCase{err: domain.NewErrUserEmailAlreadyExists("john@example.com")}
		handler := NewRegisterUserHandler(uc, newTestValidator(t))

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(validRequestBody()))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Handle(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestNewRegisterUserHandler(t *testing.T) {
	t.Run("creates handler with dependencies", func(t *testing.T) {
		uc := &mockUseCase{}
		v := newTestValidator(t)

		handler := NewRegisterUserHandler(uc, v)

		assert.NotNil(t, handler)
		assert.NotNil(t, handler.useCase)
		assert.NotNil(t, handler.validator)
	})
}

// Ensure mockUseCase implements the interface.
var _ port.RegisterUserUseCase = (*mockUseCase)(nil)
