package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"ratelab/internal/limiter"
	redisclient "ratelab/internal/redis"
)

const (
	backendURL  = "http://localhost:8081"
	gatewayPort = ":8080"
	rateLimit   = 5
	window      = 60 * time.Second
)

type Gateway struct {
	limiter   limiter.RateLimiter
	client    *http.Client
	algorithm string
}

// proxyHandler checks the rate limit before forwarding requests.
func (g *Gateway) proxyHandler(w http.ResponseWriter, r *http.Request) {
	// Identify the client for local development.
	// Production systems should use authenticated client identities.
	clientID := r.Header.Get("X-Client-ID")

	if clientID == "" {
		clientID = r.RemoteAddr
	}

	// Check the client's rate limit.
	result, err := g.limiter.Allow(r.Context(), clientID)
	if err != nil {
		fmt.Println("Rate limiter error:", err)
		http.Error(
			w,
			"rate limiter unavailable",
			http.StatusInternalServerError,
		)
		return
	}

	// Set rate-limit response headers.
	w.Header().Set(
		"X-RateLimit-Limit",
		strconv.Itoa(result.Limit),
	)
	w.Header().Set(
		"X-RateLimit-Remaining",
		strconv.Itoa(result.Remaining),
	)
	w.Header().Set(
		"X-RateLimit-Reset",
		strconv.FormatInt(result.ResetAt.Unix(), 10),
	)

	// Reject requests when the limit has been exceeded.
	if !result.Allowed {
		retryAfter := int64(time.Until(result.ResetAt).Seconds()) + 1

		if retryAfter < 1 {
			retryAfter = 1
		}

		w.Header().Set("Retry-After", strconv.FormatInt(retryAfter, 10))
		http.Error(
			w,
			`{"error":"rate limit exceeded"}`,
			http.StatusTooManyRequests,
		)
		return
	}

	// Forward allowed requests to the backend.
	targetURL := backendURL + r.URL.RequestURI()

	req, err := http.NewRequestWithContext(
		r.Context(),
		r.Method,
		targetURL,
		r.Body,
	)
	if err != nil {
		http.Error(
			w,
			"failed to create backend request",
			http.StatusInternalServerError,
		)
		return
	}

	// Copy incoming request headers.
	req.Header = r.Header.Clone()

	// Send the request to the backend.
	resp, err := g.client.Do(req)
	if err != nil {
		fmt.Println("Backend request failed:", err)
		http.Error(
			w,
			"backend unavailable",
			http.StatusBadGateway,
		)
		return
	}
	defer resp.Body.Close()

	// Copy backend response headers.
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Forward the backend response.
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func main() {
	// Connect to Redis.
	redis := redisclient.NewClient()
	defer redis.Close()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	// Verify Redis before starting the gateway.
	if err := redisclient.Ping(ctx, redis); err != nil {
		fmt.Println("Redis connection failed:", err)
		return
	}

	fmt.Println("Redis connection successful")

	// Initialize the Fixed Window rate limiter.
	var rateLimiter limiter.RateLimiter

	algorithm := os.Getenv("RATELAB_ALGORITHM")

	if algorithm == "" {
		algorithm = "sliding"
	}

	switch algorithm {
	case "fixed":
		rateLimiter = limiter.NewFixedWindow(
			redis,
			rateLimit,
			window,
		)

	case "sliding":
		rateLimiter = limiter.NewSlidingWindow(
			redis,
			rateLimit,
			window,
		)

	default:
		fmt.Println("Unknown rate-limiting algorithm:", algorithm)
		return
	}
	fmt.Println("Selected algorithm:", algorithm)

	gateway := &Gateway{
		limiter:   rateLimiter,
		client:    &http.Client{Timeout: 10 * time.Second},
		algorithm: algorithm,
	}
	// Register routes.
	mux := http.NewServeMux()
	mux.HandleFunc("/health", gateway.healthHandler)
	mux.HandleFunc("/api/", gateway.proxyHandler)

	// Start the gateway.
	fmt.Println("Gateway running on http://localhost:8080")
	fmt.Printf(
		"Rate limit: %d requests per %s per client\n",
		rateLimit,
		window,
	)

	if err := http.ListenAndServe(gatewayPort, mux); err != nil {
		fmt.Println("Gateway failed:", err)
	}
}
func (g *Gateway) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(
		w,
		`{"status":"gateway ok","algorithm":"%s"}`,
		g.algorithm,
	)
}
