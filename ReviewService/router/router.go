package router

import (
	"ReviewService/controller"

	"github.com/go-chi/chi/v5"
)

func SetupRouter() *chi.Mux {

	router := chi.NewRouter()

	router.Get("/ping", controller.PingHandler)

	return router
}
