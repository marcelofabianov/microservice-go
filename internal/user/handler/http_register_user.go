package handler

import (
	"encoding/json"
	"net/http"

	"github.com/marcelofabianov/fault"

	"github.com/marcelofabianov/course/internal/user/port"
	"github.com/marcelofabianov/course/pkg/validation"
	"github.com/marcelofabianov/course/pkg/web"
)

type RegisterUserHandler struct {
	useCase   port.RegisterUserUseCase
	validator validation.Validator
}

func NewRegisterUserHandler(useCase port.RegisterUserUseCase, validator validation.Validator) *RegisterUserHandler {
	return &RegisterUserHandler{
		useCase:   useCase,
		validator: validator,
	}
}

func (h *RegisterUserHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var input port.RegisterUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		web.Error(w, r, fault.New("invalid request body",
			fault.WithCode(fault.Invalid),
		))
		return
	}

	if err := h.validator.Struct(r.Context(), &input); err != nil {
		web.Error(w, r, err)
		return
	}

	output, err := h.useCase.Execute(r.Context(), &input)
	if err != nil {
		web.Error(w, r, err)
		return
	}

	web.Created(w, r, output.Response)
}
