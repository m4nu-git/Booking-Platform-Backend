package main

import (
	"AuthServiceInGo/app"
	dbConfig "AuthServiceInGo/config/db"
	config "AuthServiceInGo/config/env"
)


func main() {

	config.Load()
	cfg := app.NewConfig()
	app := app.NewApplication(cfg)
	dbConfig.SetupDB()
	app.Run()
}
