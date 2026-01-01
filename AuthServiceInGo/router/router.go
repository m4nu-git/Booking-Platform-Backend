package router

import (
	"AuthServiceInGo/controllers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Router interface {
	Register(r chi.Router)
}

func SetupRouter(UserRouter Router) *chi.Mux {

	chiRouter := chi.NewRouter()

	// chiRouter.Use(middlewares.RequestLogger) // Middleware for logging requests
	chiRouter.Use(middleware.Logger) // Built-in Chi middleware for logging requests

	// chiRouter.Use(middleware.RateLimitMiddleware) // Built-in Chi middlware for logging requests

	chiRouter.Get("/ping", controllers.PingHandlers)

	UserRouter.Register(chiRouter)

	return chiRouter
}