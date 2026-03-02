package middlewares

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Store IP -> limiter

var visitors = make(map[string]*rate.Limiter)
var mu sync.Mutex

// Get Limiter returns limiter for IP, or creates one
func getLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := visitors[ip]
	if !exists {
		// 5 requests per second per IP
		limiter = rate.NewLimiter(rate.Every(time.Second), 5)
		visitors[ip] = limiter
	}
	return limiter
}

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// PRODUCTION: uses the real remote address of the client
		// ip, _, err := net.SplitHostPort(r.RemoteAddr)
		// if err != nil {
		// 	http.Error(w, "Unable to determine IP", http.StatusInternalServerError)
		// 	return
		// }

		// TESTING ONLY: uncomment below (and comment the block above) to simulate
		// different IPs via the X-Forwarded-For header in Postman.
		// In Postman → Headers → add: X-Forwarded-For : <any-fake-ip>
		// Each unique IP value gets its own rate limit bucket.

		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip, _, _ = net.SplitHostPort(r.RemoteAddr)
		}

		limiter := getLimiter(ip)
		if !limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
