package main

import (
	"fmt"
	"io"
	"net/http"
)

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	backendURL := "http://localhost:8081" + r.URL.Path

	req, err := http.NewRequest(r.Method, backendURL, r.Body)
	if err != nil {
		http.Error(w, "failed to create backend request", http.StatusInternalServerError)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "backend unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	io.Copy(w, resp.Body)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"status":"gateway ok"}`)
}

func main() {
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/api/", proxyHandler)

	fmt.Println("Gateway running on http://localhost:8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Gateway failed:", err)
	}
}
