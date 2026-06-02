package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// ProxyToService returns an http.HandlerFunc that reverse-proxies requests to
// targetBaseUrl after stripping pathPrefix from the incoming request path.
//
// It also:
//   - Injects the authenticated user's ID as X-User-ID for downstream services.
//   - Forwards x-correlation-id so distributed traces remain intact.
//   - Sets a ResponseHeaderTimeout so the gateway never hangs on a slow upstream.
//   - Returns a clean JSON error envelope when the upstream is unreachable.
func ProxyToService(targetBaseUrl string, pathPrefix string) http.HandlerFunc {

	target, err := url.Parse(targetBaseUrl)
	if err != nil {
		fmt.Println("Error parsing target URL:", err)
		return nil
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Prevent the gateway from hanging indefinitely on a slow or dead upstream.
	proxy.Transport = &http.Transport{
		ResponseHeaderTimeout: 10 * time.Second,
	}

	originalDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		originalDirector(r)

		// Strip the service namespace prefix, then forward the remainder to the upstream.
		// Example: /hotelService/api/v1/hotels  →  /api/v1/hotels
		strippedPath := strings.TrimPrefix(r.URL.Path, pathPrefix)
		r.URL.Host = target.Host
		r.URL.Path = target.Path + strippedPath
		r.Host = target.Host

		// Forward the authenticated user's identity so downstream services can
		// apply per-user logic without re-validating the JWT themselves.
		if userID, ok := r.Context().Value("userID").(string); ok {
			r.Header.Set("X-User-ID", userID)
		}
		if email, ok := r.Context().Value("email").(string); ok {
			r.Header.Set("X-User-Email", email)
		}

		// Propagate the correlation ID so log lines across all services can be
		// joined by a single trace ID.
		if corrID := r.Header.Get("x-correlation-id"); corrID != "" {
			r.Header.Set("x-correlation-id", corrID)
		}
	}

	// When the upstream returns a 5xx, ensure the Content-Type is JSON so the
	// client always receives a parseable response regardless of the upstream's format.
	proxy.ModifyResponse = func(resp *http.Response) error {
		if resp.StatusCode >= 500 {
			resp.Header.Set("Content-Type", "application/json")
		}
		return nil
	}

	// Convert network-level proxy errors (upstream unreachable, timeout, etc.)
	// into the project's standard JSON error envelope.
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		fmt.Printf("Gateway proxy error [%s %s]: %v\n", r.Method, r.URL.Path, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "upstream service unavailable",
			"error":   err.Error(),
		})
	}

	return proxy.ServeHTTP
}
