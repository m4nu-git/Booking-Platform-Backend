package main

import (
	"AuthServiceInGo/app"
	config "AuthServiceInGo/config/env"
)


func main() {

	config.Load()
	cfg := app.NewConfig()
	app := app.NewApplication(cfg)

	app.Run()
}
