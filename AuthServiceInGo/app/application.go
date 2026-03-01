package app

import (
	dbConfig "AuthServiceInGo/config/db"
	config "AuthServiceInGo/config/env"
	"AuthServiceInGo/controllers"
	repo "AuthServiceInGo/db/repositories"
	"AuthServiceInGo/router"
	"AuthServiceInGo/services"
	"fmt"
	"net/http"
	"time"
)

// Config holds the configuration for the server.
type Config struct {
	Addr string // port
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

	db, err := dbConfig.SetupDB()

	if err != nil {
		fmt.Println("Error setting up database:", err)
		return err
	}

	ur := repo.NewUserRepository(db)
	rr := repo.NewRoleRepository(db)
	rpr := repo.NewRolePermissionRepository(db)
	urr := repo.NewUserRoleRepository(db)
	pr := repo.NewPermissionRepository(db)
	us := services.NewUserService(ur, rr, urr)
	rs := services.NewRoleService(rr, rpr, urr)
	ps := services.NewPermissionService(pr)
	uc := controllers.NewUserController(us)
	rc := controllers.NewRoleController(rs)
	pc := controllers.NewPermissionController(ps)
	uRouter := router.NewUserRouter(uc)
	rRouter := router.NewRoleRouter(rc)
	pRouter := router.NewPermissionRouter(pc)

	server := &http.Server{
		Addr:         app.Config.Addr,
		Handler:      router.SetupRouter(uRouter, rRouter, pRouter),
		ReadTimeout:  10 * time.Second,                     // Set read timeout to 10 seconds
		WriteTimeout: 10 * time.Second,                     // Set write timeout to 10 seconds
	}

	fmt.Println("Starting server on", app.Config.Addr)

	return server.ListenAndServe()
}
