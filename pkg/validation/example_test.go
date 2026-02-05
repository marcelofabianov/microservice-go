package validation_test

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/marcelofabianov/course/pkg/logger"
	"github.com/marcelofabianov/course/pkg/validation"
)

func Example_basicValidation() {
	log := logger.New(&logger.Config{
		Level:       logger.LevelInfo,
		Format:      logger.FormatText,
		ServiceName: "example",
		Environment: "development",
	})

	v := validation.New(log, nil)
	ctx := context.Background()

	type User struct {
		Name  string `json:"name" validate:"required,min=3"`
		Email string `json:"email" validate:"required,email"`
		Age   int    `json:"age" validate:"required,gte=18"`
	}

	user := User{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   25,
	}

	if err := v.Struct(ctx, user); err != nil {
		log.Error("Validation failed", "error", err)
		return
	}

	fmt.Println("User validated successfully")
	// Output: User validated successfully
}

func Example_fieldValidation() {
	log := logger.New(&logger.Config{
		Level:       logger.LevelInfo,
		Format:      logger.FormatText,
		ServiceName: "example",
		Environment: "development",
	})

	v := validation.New(log, nil)
	ctx := context.Background()

	email := "john@example.com"

	if err := v.Field(ctx, email, "required,email"); err != nil {
		log.Error("Email validation failed", "error", err)
		return
	}

	fmt.Println("Email validated successfully")
	// Output: Email validated successfully
}

func Example_brazilianValidation() {
	log := logger.New(&logger.Config{
		Level:       logger.LevelInfo,
		Format:      logger.FormatText,
		ServiceName: "example",
		Environment: "development",
	})

	v := validation.New(log, nil)

	if err := validation.RegisterBrazilianValidators(v); err != nil {
		panic(err)
	}

	ctx := context.Background()

	type BrazilianUser struct {
		Name  string `json:"name" validate:"required"`
		CPF   string `json:"cpf" validate:"required,cpf"`
		Phone string `json:"phone" validate:"required,phone"`
		CEP   string `json:"cep" validate:"required,cep"`
	}

	user := BrazilianUser{
		Name:  "João Silva",
		CPF:   "123.456.789-09",
		Phone: "(11) 91234-5678",
		CEP:   "12345-678",
	}

	if err := v.Struct(ctx, user); err != nil {
		log.Error("Validation failed", "error", err)
		return
	}

	fmt.Println("Brazilian user validated successfully")
	// Output: Brazilian user validated successfully
}

func Example_customValidator() {
	log := logger.New(&logger.Config{
		Level:       logger.LevelInfo,
		Format:      logger.FormatText,
		ServiceName: "example",
		Environment: "development",
	})

	v := validation.New(log, nil)
	ctx := context.Background()

	err := v.RegisterCustom("is_even", func(fl validator.FieldLevel) bool {
		value := fl.Field().Int()
		return value%2 == 0
	})
	if err != nil {
		panic(err)
	}

	type TestStruct struct {
		Number int `json:"number" validate:"required,is_even"`
	}

	data := TestStruct{Number: 4}

	if err := v.Struct(ctx, data); err != nil {
		log.Error("Validation failed", "error", err)
		return
	}

	fmt.Println("Custom validation passed")
	// Output: Custom validation passed
}

func Example_sanitizeSensitiveData() {
	log := logger.New(&logger.Config{
		Level:       logger.LevelError,
		Format:      logger.FormatText,
		ServiceName: "example",
		Environment: "development",
	})

	cfg := &validation.Config{
		EnableLogging:         false,
		SanitizeSensitiveData: true,
	}

	v := validation.New(log, cfg)
	ctx := context.Background()

	type LoginForm struct {
		Username string `json:"username" validate:"required"`
		Password string `json:"password" validate:"required,min=8"`
	}

	form := LoginForm{
		Username: "john",
		Password: "short",
	}

	err := v.Struct(ctx, form)
	if err != nil {
		fmt.Println("Password field will be redacted in logs")
	}

	// Output: Password field will be redacted in logs
}

func Example_complexStructValidation() {
	log := logger.New(&logger.Config{
		Level:       logger.LevelInfo,
		Format:      logger.FormatText,
		ServiceName: "example",
		Environment: "development",
	})

	v := validation.New(log, nil)
	ctx := context.Background()

	type Address struct {
		Street  string `json:"street" validate:"required"`
		Number  int    `json:"number" validate:"required,gt=0"`
		ZipCode string `json:"zip_code" validate:"required"`
		City    string `json:"city" validate:"required"`
	}

	type Person struct {
		Name    string  `json:"name" validate:"required,min=3"`
		Email   string  `json:"email" validate:"required,email"`
		Age     int     `json:"age" validate:"required,gte=18"`
		Address Address `json:"address" validate:"required"`
	}

	person := Person{
		Name:  "Jane Doe",
		Email: "jane@example.com",
		Age:   30,
		Address: Address{
			Street:  "Main Street",
			Number:  123,
			ZipCode: "12345",
			City:    "São Paulo",
		},
	}

	if err := v.Struct(ctx, person); err != nil {
		log.Error("Validation failed", "error", err)
		return
	}

	fmt.Println("Complex struct validated successfully")
	// Output: Complex struct validated successfully
}

func Example_httpRequestValidation() {
	log := logger.New(&logger.Config{
		Level:       logger.LevelInfo,
		Format:      logger.FormatText,
		ServiceName: "api",
		Environment: "production",
	})

	v := validation.New(log, nil)

	if err := validation.RegisterBrazilianValidators(v); err != nil {
		panic(err)
	}

	type CreateUserRequest struct {
		Name     string `json:"name" validate:"required,min=3,max=100"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
		CPF      string `json:"cpf" validate:"required,cpf"`
		Phone    string `json:"phone" validate:"required,phone"`
		Age      int    `json:"age" validate:"required,gte=18,lte=120"`
	}

	ctx := context.Background()
	request := CreateUserRequest{
		Name:     "Maria Silva",
		Email:    "maria@example.com",
		Password: "SecurePass123",
		CPF:      "123.456.789-09",
		Phone:    "(11) 91234-5678",
		Age:      28,
	}

	if err := v.Struct(ctx, request); err != nil {
		fmt.Println("Request validation failed")
		return
	}

	fmt.Println("HTTP request validated successfully")
	// Output: HTTP request validated successfully
}
