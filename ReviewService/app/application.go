package app

import (
	config "ReviewService/config/env"
	"ReviewService/router"
	"fmt"
	"net/http"
	"time"
)

type Config struct {
	Addr string //PORT
}

type Application struct {
	Config Config
}

func NewConfig() Config {

	port := config.GetString("PORT", ":5555")

	return Config{
		Addr: port,
	}
}

func NewApplication(cfg Config) *Application {
	return &Application{
		Config: cfg,
	}
}

func (app *Application) Run() error {

	server := &http.Server{
		Addr:         app.Config.Addr,
		Handler:      router.SetupRouter(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("Starting server on", app.Config.Addr)

	return server.ListenAndServe()
}
