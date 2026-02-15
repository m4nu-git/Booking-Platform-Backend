package main

import (
	"ReviewService/app"
)

func main() {

	cfg := app.NewConfig(":4000")
	app := app.NewApplication(cfg)

	app.Run()
}
