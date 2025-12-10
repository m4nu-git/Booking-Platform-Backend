package app

import (
	"fmt"
	"net/http"
	"time"
)

type Config struct {
	Addr string
}

func NewConfig(addr string) Config {
	return Config{
		Addr: addr,
	}
}


type Application struct {
	Config Config
}

func NewApplication(cfg Config) *Application {
	return &Application{
		Config: cfg,
	}
}


func (app *Application) Run() error {
	server := &http.Server{
		Addr: app.Config.Addr,
		Handler: nil, // TODO: Setup a chi router and put it here
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("Starting server on", app.Config.Addr)

	return server.ListenAndServe()
}