package gateway

import "net/http"

func (g *Gateway) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", g.healthHandler)

	mux.Handle("/api/", http.HandlerFunc(g.proxyHandler))

	return mux
}
