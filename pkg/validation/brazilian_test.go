package validation_test

import (
	"context"
	"testing"

	"github.com/marcelofabianov/course/pkg/logger"
	"github.com/marcelofabianov/course/pkg/validation"
)

func TestBrazilianValidators(t *testing.T) {
	log := logger.New(&logger.Config{
		Level:       logger.LevelInfo,
		Format:      logger.FormatText,
		ServiceName: "test",
		Environment: "test",
	})

	v := validation.New(log, nil)
	ctx := context.Background()

	err := validation.RegisterBrazilianValidators(v)
	if err != nil {
		t.Fatalf("failed to register brazilian validators: %v", err)
	}

	t.Run("validates CPF", func(t *testing.T) {
		type TestStruct struct {
			CPF string `json:"cpf" validate:"cpf"`
		}

		tests := []struct {
			name    string
			cpf     string
			wantErr bool
		}{
			{"valid CPF with mask", "123.456.789-09", false},
			{"valid CPF without mask", "12345678909", false},
			{"invalid CPF", "123.456.789-00", true},
			{"empty CPF", "", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				data := TestStruct{CPF: tt.cpf}
				err := v.Struct(ctx, data)
				if (err != nil) != tt.wantErr {
					t.Errorf("CPF validation error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("validates CNPJ", func(t *testing.T) {
		type TestStruct struct {
			CNPJ string `json:"cnpj" validate:"cnpj"`
		}

		tests := []struct {
			name    string
			cnpj    string
			wantErr bool
		}{
			{"valid CNPJ with mask", "11.222.333/0001-81", false},
			{"valid CNPJ without mask", "11222333000181", false},
			{"invalid CNPJ", "11.222.333/0001-00", true},
			{"empty CNPJ", "", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				data := TestStruct{CNPJ: tt.cnpj}
				err := v.Struct(ctx, data)
				if (err != nil) != tt.wantErr {
					t.Errorf("CNPJ validation error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("validates CEP", func(t *testing.T) {
		type TestStruct struct {
			CEP string `json:"cep" validate:"cep"`
		}

		tests := []struct {
			name    string
			cep     string
			wantErr bool
		}{
			{"valid CEP with mask", "12345-678", false},
			{"valid CEP without mask", "12345678", false},
			{"invalid CEP", "1234", true},
			{"empty CEP", "", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				data := TestStruct{CEP: tt.cep}
				err := v.Struct(ctx, data)
				if (err != nil) != tt.wantErr {
					t.Errorf("CEP validation error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("validates Phone", func(t *testing.T) {
		type TestStruct struct {
			Phone string `json:"phone" validate:"phone"`
		}

		tests := []struct {
			name    string
			phone   string
			wantErr bool
		}{
			{"valid mobile with mask", "(11) 91234-5678", false},
			{"valid mobile without mask", "11912345678", false},
			{"invalid phone", "123", true},
			{"empty phone", "", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				data := TestStruct{Phone: tt.phone}
				err := v.Struct(ctx, data)
				if (err != nil) != tt.wantErr {
					t.Errorf("Phone validation error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("validates Email with wisp", func(t *testing.T) {
		type TestStruct struct {
			Email string `json:"email" validate:"email"`
		}

		tests := []struct {
			name    string
			email   string
			wantErr bool
		}{
			{"valid email", "john@example.com", false},
			{"valid email with subdomain", "john@mail.example.com", false},
			{"invalid email", "invalid-email", true},
			{"empty email", "", false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				data := TestStruct{Email: tt.email}
				err := v.Struct(ctx, data)
				if (err != nil) != tt.wantErr {
					t.Errorf("Email validation error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})
}

func TestBrazilianValidators_CompleteForm(t *testing.T) {
	log := logger.New(&logger.Config{
		Level:       logger.LevelInfo,
		Format:      logger.FormatText,
		ServiceName: "test",
		Environment: "test",
	})

	v := validation.New(log, nil)
	ctx := context.Background()

	err := validation.RegisterBrazilianValidators(v)
	if err != nil {
		t.Fatalf("failed to register brazilian validators: %v", err)
	}

	type BrazilianForm struct {
		Name  string `json:"name" validate:"required,min=3"`
		Email string `json:"email" validate:"required,email"`
		CPF   string `json:"cpf" validate:"required,cpf"`
		Phone string `json:"phone" validate:"required,phone"`
		CEP   string `json:"cep" validate:"required,cep"`
	}

	t.Run("valid complete form", func(t *testing.T) {
		form := BrazilianForm{
			Name:  "João Silva",
			Email: "joao@example.com",
			CPF:   "123.456.789-09",
			Phone: "(11) 91234-5678",
			CEP:   "12345-678",
		}

		err := v.Struct(ctx, form)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("invalid complete form", func(t *testing.T) {
		form := BrazilianForm{
			Name:  "Jo",
			Email: "invalid",
			CPF:   "000.000.000-00",
			Phone: "123",
			CEP:   "123",
		}

		err := v.Struct(ctx, form)
		if err == nil {
			t.Fatal("expected error for invalid form, got nil")
		}
	})
}

func TestBrazilianValidators_Registration(t *testing.T) {
	log := logger.New(&logger.Config{
		Level:       logger.LevelInfo,
		Format:      logger.FormatText,
		ServiceName: "test",
		Environment: "test",
	})

	v := validation.New(log, nil)

	t.Run("registers all validators successfully", func(t *testing.T) {
		err := validation.RegisterBrazilianValidators(v)
		if err != nil {
			t.Fatalf("failed to register validators: %v", err)
		}
	})

	t.Run("can register multiple times", func(t *testing.T) {
		v2 := validation.New(log, nil)
		err := validation.RegisterBrazilianValidators(v2)
		if err != nil {
			t.Fatalf("failed to register validators: %v", err)
		}
	})
}
