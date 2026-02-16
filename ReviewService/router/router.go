package router

import (
	"ReviewService/controllers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Router interface {
	Register(r chi.Router)
}

func SetupRouter(ReviewRouter Router) *chi.Mux {

	chiRouter := chi.NewRouter()

	chiRouter.Use(middleware.Logger)

	chiRouter.Get("/ping", controllers.PingHandler)

	ReviewRouter.Register(chiRouter)

	return chiRouter
}
