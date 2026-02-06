package handler

import "github.com/go-chi/chi/v5"

type UserRouter struct {
	registerHandler *RegisterUserHandler
}

func NewUserRouter(registerHandler *RegisterUserHandler) *UserRouter {
	return &UserRouter{registerHandler: registerHandler}
}

func (ur *UserRouter) RegisterRoutes(r chi.Router) {
	r.Route("/users", func(r chi.Router) {
		r.Post("/", ur.registerHandler.Handle)
	})
}
