package gateway

import (
	"AuthServiceInGo/middlewares"
	"AuthServiceInGo/utils"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewGatewayRouter() http.Handler {
	r := chi.NewRouter()

	// Forward requests to HotelService
	r.With(middlewares.JWTAuthMiddleware, middlewares.RequireAllRoles("user", "admin")).HandleFunc("/hotelService/*", utils.ProxyToService(
		"http://localhost:3000",
		"/hotelService",
	))

	// Forward requests to BookingService
	r.With(middlewares.JWTAuthMiddleware, middlewares.RequireAnyRole("user", "admin")).HandleFunc("/bookingService/*", utils.ProxyToService(
		"http://localhost:3005",
		"/bookingService",
	))

	r.With(middlewares.JWTAuthMiddleware, middlewares.RequireAnyRole("user", "admin")).HandleFunc("/reviewService/*", utils.ProxyToService(
		"http://localhost:8081",
		"/reviewService",
	))

	return r
}
