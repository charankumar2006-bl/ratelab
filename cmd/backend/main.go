package main

import (
	"fmt"
	"net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	fmt.Fprint(w, `{"status":"ok"}`)
}

func main() {
	http.HandleFunc("/health", healthHandler)

	fmt.Println("Backend server running on http://localhost:8081")

	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		fmt.Println("Server failed:", err)
	}
}
