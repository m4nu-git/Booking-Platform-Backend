package app

import (
	client "ReviewService/client"
	dbConfig "ReviewService/config/db"
	config "ReviewService/config/env"
	"ReviewService/controllers"
	cronjob "ReviewService/cronJob"
	repo "ReviewService/db/repositories"
	"ReviewService/router"
	"ReviewService/services"
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

	db, err := dbConfig.SetupDB()

	if err != nil {
		fmt.Println("Error settings up database:", err)
		return err
	}

	rr := repo.NewReviewRepository(db)
	rs := services.NewReviewService(rr)
	rc := controllers.NewReviewController(rs)
	rRouter := router.NewReviewRouter(rc)

	server := &http.Server{
		Addr:         app.Config.Addr,
		Handler:      router.SetupRouter(rRouter),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("Starting Review Service on", app.Config.Addr)

	// Create dependent Services

	repository := repo.NewReviewAggregateRatingRepository(db)
	hotelClient := client.NewHotelClient(config.GetString("HOTEL_SERVICE_URL", "http://localhost:3000/api/v1"))
	svc := services.NewReviewBatchProcessor(db, repository, hotelClient)

	// start cron job
	mode := config.GetString("APP_MODE", "test")
	cronjob.StartCron(svc, mode)

	return server.ListenAndServe()
}
