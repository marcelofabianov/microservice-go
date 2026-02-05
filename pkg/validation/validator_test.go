package validation_test

import (
	"context"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/marcelofabianov/course/pkg/logger"
	"github.com/marcelofabianov/course/pkg/validation"
)

func TestNew(t *testing.T) {
	t.Run("creates validator with default config", func(t *testing.T) {
		log := logger.New(&logger.Config{
			Level:       logger.LevelInfo,
			Format:      logger.FormatText,
			ServiceName: "test",
			Environment: "test",
		})

		v := validation.New(log, nil)
		if v == nil {
			t.Fatal("expected validator, got nil")
		}
	})

	t.Run("creates validator with custom config", func(t *testing.T) {
		log := logger.New(&logger.Config{
			Level:       logger.LevelInfo,
			Format:      logger.FormatText,
			ServiceName: "test",
			Environment: "test",
		})

		cfg := &validation.Config{
			EnableLogging:         false,
			SanitizeSensitiveData: false,
		}

		v := validation.New(log, cfg)
		if v == nil {
			t.Fatal("expected validator, got nil")
		}
	})
}

func TestValidator_Struct(t *testing.T) {
	log := logger.New(&logger.Config{
		Level:       logger.LevelInfo,
		Format:      logger.FormatText,
		ServiceName: "test",
		Environment: "test",
	})

	v := validation.New(log, nil)
	ctx := context.Background()

	t.Run("valid struct passes validation", func(t *testing.T) {
		type TestStruct struct {
			Name  string `json:"name" validate:"required,min=3"`
			Email string `json:"email" validate:"required,email"`
			Age   int    `json:"age" validate:"required,gte=18"`
		}

		data := TestStruct{
			Name:  "John Doe",
			Email: "john@example.com",
			Age:   25,
		}

		err := v.Struct(ctx, data)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("invalid struct fails validation", func(t *testing.T) {
		type TestStruct struct {
			Name  string `json:"name" validate:"required,min=3"`
			Email string `json:"email" validate:"required,email"`
		}

		data := TestStruct{
			Name:  "Jo",
			Email: "invalid-email",
		}

		err := v.Struct(ctx, data)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("nil struct returns error", func(t *testing.T) {
		err := v.Struct(ctx, nil)
		if err == nil {
			t.Fatal("expected error for nil struct, got nil")
		}
	})

	t.Run("sanitizes sensitive fields", func(t *testing.T) {
		type TestStruct struct {
			Username string `json:"username" validate:"required"`
			Password string `json:"password" validate:"required,min=8"`
		}

		data := TestStruct{
			Username: "john",
			Password: "short",
		}

		err := v.Struct(ctx, data)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestValidator_Field(t *testing.T) {
	log := logger.New(&logger.Config{
		Level:       logger.LevelInfo,
		Format:      logger.FormatText,
		ServiceName: "test",
		Environment: "test",
	})

	v := validation.New(log, nil)
	ctx := context.Background()

	t.Run("valid field passes validation", func(t *testing.T) {
		email := "john@example.com"
		err := v.Field(ctx, email, "required,email")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("invalid field fails validation", func(t *testing.T) {
		email := "invalid-email"
		err := v.Field(ctx, email, "required,email")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("nil field returns error", func(t *testing.T) {
		err := v.Field(ctx, nil, "required")
		if err == nil {
			t.Fatal("expected error for nil field, got nil")
		}
	})

	t.Run("empty tag returns error", func(t *testing.T) {
		err := v.Field(ctx, "value", "")
		if err == nil {
			t.Fatal("expected error for empty tag, got nil")
		}
	})
}

func TestValidator_RegisterCustom(t *testing.T) {
	log := logger.New(&logger.Config{
		Level:       logger.LevelInfo,
		Format:      logger.FormatText,
		ServiceName: "test",
		Environment: "test",
	})

	v := validation.New(log, nil)

	t.Run("registers custom validator successfully", func(t *testing.T) {
		err := v.RegisterCustom("custom", func(fl validator.FieldLevel) bool {
			return true
		})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("empty tag returns error", func(t *testing.T) {
		err := v.RegisterCustom("", func(fl validator.FieldLevel) bool {
			return true
		})
		if err == nil {
			t.Fatal("expected error for empty tag, got nil")
		}
	})

	t.Run("nil function returns error", func(t *testing.T) {
		err := v.RegisterCustom("custom", nil)
		if err == nil {
			t.Fatal("expected error for nil function, got nil")
		}
	})
}

func TestConfig_Default(t *testing.T) {
	cfg := validation.DefaultConfig()

	if !cfg.EnableLogging {
		t.Error("expected EnableLogging to be true")
	}

	if !cfg.SanitizeSensitiveData {
		t.Error("expected SanitizeSensitiveData to be true")
	}

	if cfg.LogSuccessfulValidations {
		t.Error("expected LogSuccessfulValidations to be false")
	}

	if len(cfg.AdditionalSensitiveFields) != 0 {
		t.Error("expected AdditionalSensitiveFields to be empty")
	}
}

func TestSensitiveDataSanitization(t *testing.T) {
	log := logger.New(&logger.Config{
		Level:       logger.LevelInfo,
		Format:      logger.FormatText,
		ServiceName: "test",
		Environment: "test",
	})

	cfg := &validation.Config{
		EnableLogging:             true,
		SanitizeSensitiveData:     true,
		AdditionalSensitiveFields: []string{"custom_secret"},
	}

	v := validation.New(log, cfg)
	ctx := context.Background()

	type TestStruct struct {
		Username     string `json:"username" validate:"required"`
		Password     string `json:"password" validate:"required,min=8"`
		Token        string `json:"token" validate:"required"`
		CustomSecret string `json:"custom_secret" validate:"required"`
	}

	data := TestStruct{
		Username:     "john",
		Password:     "short",
		Token:        "tok",
		CustomSecret: "sec",
	}

	err := v.Struct(ctx, data)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestValidation_ComplexStruct(t *testing.T) {
	log := logger.New(&logger.Config{
		Level:       logger.LevelInfo,
		Format:      logger.FormatText,
		ServiceName: "test",
		Environment: "test",
	})

	v := validation.New(log, nil)
	ctx := context.Background()

	type Address struct {
		Street  string `json:"street" validate:"required"`
		Number  int    `json:"number" validate:"required,gt=0"`
		ZipCode string `json:"zip_code" validate:"required"`
	}

	type User struct {
		Name    string  `json:"name" validate:"required,min=3,max=100"`
		Email   string  `json:"email" validate:"required,email"`
		Age     int     `json:"age" validate:"required,gte=18,lte=120"`
		Address Address `json:"address" validate:"required"`
	}

	t.Run("valid complex struct", func(t *testing.T) {
		user := User{
			Name:  "John Doe",
			Email: "john@example.com",
			Age:   30,
			Address: Address{
				Street:  "Main Street",
				Number:  123,
				ZipCode: "12345",
			},
		}

		err := v.Struct(ctx, user)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("invalid nested struct", func(t *testing.T) {
		user := User{
			Name:  "John Doe",
			Email: "john@example.com",
			Age:   30,
			Address: Address{
				Street:  "",
				Number:  0,
				ZipCode: "12345",
			},
		}

		err := v.Struct(ctx, user)
		if err == nil {
			t.Fatal("expected error for invalid nested struct, got nil")
		}
	})
}
