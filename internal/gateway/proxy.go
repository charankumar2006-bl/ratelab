package gateway

import (
	"log/slog"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (g *Gateway) proxyHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	// X-Client-ID is intended for local testing.
	// Production systems should use a trusted client identity.
	clientID := strings.TrimSpace(r.Header.Get("X-Client-ID"))

	if clientID == "" {
		clientID = clientAddress(r.RemoteAddr)
	}

	result, err := g.limiter.Allow(r.Context(), clientID)
	if err != nil {
		slog.Error("rate limiter failed", "error", err)

		writeJSONError(
			w,
			http.StatusServiceUnavailable,
			"rate limiter unavailable",
		)
		return
	}

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

	if !result.Allowed {
		retryAfter := int64(
			math.Ceil(time.Until(result.ResetAt).Seconds()),
		)

		if retryAfter < 1 {
			retryAfter = 1
		}

		w.Header().Set(
			"Retry-After",
			strconv.FormatInt(retryAfter, 10),
		)

		writeJSONError(
			w,
			http.StatusTooManyRequests,
			"rate limit exceeded",
		)
		return
	}

	g.proxy.ServeHTTP(w, r)
}

func clientAddress(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		return host
	}

	return remoteAddr
}
