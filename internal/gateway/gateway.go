package gateway

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"ratelab/internal/limiter"
)

type Gateway struct {
	limiter   limiter.RateLimiter
	algorithm string
	proxy     *httputil.ReverseProxy
}

func New(
	rateLimiter limiter.RateLimiter,
	backendURL string,
	algorithm string,
) (*Gateway, error) {
	if rateLimiter == nil {
		return nil, fmt.Errorf("rate limiter must not be nil")
	}

	target, err := url.Parse(backendURL)
	if err != nil {
		return nil, fmt.Errorf("parse backend URL: %w", err)
	}

	if (target.Scheme != "http" && target.Scheme != "https") ||
		target.Host == "" {
		return nil, fmt.Errorf("backend URL must be an absolute HTTP or HTTPS URL")
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	if transport, ok := http.DefaultTransport.(*http.Transport); ok {
		customTransport := transport.Clone()
		customTransport.ResponseHeaderTimeout = 10 * time.Second
		proxy.Transport = customTransport
	}

	return &Gateway{
		limiter:   rateLimiter,
		algorithm: algorithm,
		proxy:     proxy,
	}, nil
}
