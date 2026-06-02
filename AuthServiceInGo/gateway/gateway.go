package gateway

import (
	config "AuthServiceInGo/config/env"
	"AuthServiceInGo/middlewares"
	"AuthServiceInGo/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// corsMiddleware sets CORS and security headers on every proxied response.
// The allowed origin is configurable via CORS_ALLOWED_ORIGIN env var.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := config.GetString("CORS_ALLOWED_ORIGIN", "*")

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Correlation-ID")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Security hardening headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")

		// Preflight requests are handled here — no need to hit upstream
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type serviceStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Latency string `json:"latency"`
}

// pingService hits a single downstream health endpoint and reports its status.
func pingService(name, baseURL, pingPath string, ch chan<- serviceStatus) {
	start := time.Now()
	client := &http.Client{Timeout: 3 * time.Second}

	resp, err := client.Get(baseURL + pingPath)
	latency := fmt.Sprintf("%dms", time.Since(start).Milliseconds())

	if err != nil {
		ch <- serviceStatus{Name: name, Status: "down", Latency: latency}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		ch <- serviceStatus{Name: name, Status: "down", Latency: latency}
		return
	}

	ch <- serviceStatus{Name: name, Status: "up", Latency: latency}
}

// gatewayHealthHandler pings all 4 downstream services in parallel and returns
// a JSON summary. Returns 503 if any service is down.
func gatewayHealthHandler(hotelURL, bookURL, notifURL, reviewURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ch := make(chan serviceStatus, 4)

		// ReviewService uses /ping (no /api/v1 prefix); others use /api/v1/ping/health
		go pingService("hotel-service",        hotelURL,  "/api/v1/ping/health", ch)
		go pingService("booking-service",      bookURL,   "/api/v1/ping/health", ch)
		go pingService("notification-service", notifURL,  "/api/v1/ping/health", ch)
		go pingService("review-service",       reviewURL, "/ping",               ch)

		results := make([]serviceStatus, 0, 4)
		for i := 0; i < 4; i++ {
			results = append(results, <-ch)
		}

		allUp := true
		for _, s := range results {
			if s.Status != "up" {
				allUp = false
				break
			}
		}

		overall := "healthy"
		statusCode := http.StatusOK
		if !allUp {
			overall = "degraded"
			statusCode = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"gateway":  "up",
			"overall":  overall,
			"services": results,
		})
	}
}

// NewGatewayRouter builds the reverse-proxy sub-router. It is mounted onto the
// main chi router in app/application.go so that global middleware (rate limiting,
// request logging) applies before these routes are reached.
//
// Route → Upstream mapping:
//
//	/hotelService/*        → HotelService       (HOTEL_SERVICE_URL, :3000)  – any authenticated user
//	/bookingService/*      → BookingService      (BOOKING_SERVICE_URL, :3001) – any authenticated user
//	/notificationService/* → NotificationService (NOTIFICATION_SERVICE_URL, :3002) – admin only
//	/reviewService/*       → ReviewService       (REVIEW_SERVICE_URL, :4000)  – any authenticated user
//
// Path stripping example:
//
//	GET /hotelService/api/v1/hotels  →  GET http://localhost:3000/api/v1/hotels
func NewGatewayRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(corsMiddleware)

	hotelURL  := config.GetString("HOTEL_SERVICE_URL",        "http://localhost:3000")
	bookURL   := config.GetString("BOOKING_SERVICE_URL",      "http://localhost:3001")
	notifURL  := config.GetString("NOTIFICATION_SERVICE_URL", "http://localhost:3002")
	reviewURL := config.GetString("REVIEW_SERVICE_URL",       "http://localhost:4000")

	auth      := middlewares.JWTAuthMiddleware
	anyRole   := middlewares.RequireAnyRole("user", "admin")
	adminOnly := middlewares.RequireAllRoles("admin")

	// Health check — unauthenticated, aggregates all downstream service statuses
	r.Get("/health", gatewayHealthHandler(hotelURL, bookURL, notifURL, reviewURL))

	// Proxy routes — all require a valid JWT
	r.With(auth, anyRole).HandleFunc("/hotelService/*",
		utils.ProxyToService(hotelURL, "/hotelService"))

	r.With(auth, anyRole).HandleFunc("/bookingService/*",
		utils.ProxyToService(bookURL, "/bookingService"))

	// NotificationService is an internal service — restrict to admins only
	r.With(auth, adminOnly).HandleFunc("/notificationService/*",
		utils.ProxyToService(notifURL, "/notificationService"))

	r.With(auth, anyRole).HandleFunc("/reviewService/*",
		utils.ProxyToService(reviewURL, "/reviewService"))

	return r
}
