package main

import (
	"ReviewService/app"
	config "ReviewService/config/env"
)

func main() {

	config.Load()
	cfg := app.NewConfig()
	app := app.NewApplication(cfg)

	app.Run()
}
