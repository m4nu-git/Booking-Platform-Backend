package app

import (
	config "AuthServiceInGo/config/env"
	"AuthServiceInGo/controllers"
	db "AuthServiceInGo/db/repositories"
	"AuthServiceInGo/router"
	"AuthServiceInGo/services"
	"fmt"
	"net/http"
	"time"
)

type Config struct {
	Addr string
}

func NewConfig() Config {
	port := config.GetString("PORT", ":8080")
	return Config{
		Addr: port,
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

	ur := db.NewUserRespository()
	us := services.NewUserService(ur)
	uc := controllers.NewUserController(us)
	uRouter := router.NewUserRouter(uc)


	server := &http.Server{
		Addr: app.Config.Addr,
		Handler: router.SetupRouter(uRouter), // TODO: Setup a chi router and put it here
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("Starting server on", app.Config.Addr)

	return server.ListenAndServe()
}