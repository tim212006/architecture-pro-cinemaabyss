package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
)

func main() {
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/api/movies", moviesHandler)
	http.HandleFunc("/api/users", usersHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	address := fmt.Sprintf(":%s", port)
	log.Printf("Starting proxy service on %s", address)

	log.Fatal(http.ListenAndServe(address, nil))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not Allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Strangler Fig Proxy is healthy"))
}

func shouldUseMoviesService() bool {
	if os.Getenv("GRADUAL_MIGRATION") != "true" {
		return false
	}

	percent, err := strconv.Atoi(os.Getenv("MOVIES_MIGRATION_PERCENT"))
	if err != nil {
		percent = 0
	}

	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	return rand.Intn(100) < percent
}

func moviesHandler(w http.ResponseWriter, r *http.Request) {
	target := os.Getenv("MONOLITH_URL")
	if shouldUseMoviesService() {
		target = os.Getenv("MOVIES_SERVICE_URL")
	}

	proxyRequest(w, r, target)
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	proxyRequest(w, r, os.Getenv("MONOLITH_URL"))
}

func proxyRequest(w http.ResponseWriter, r *http.Request, target string) {
	targetURL, err := url.Parse(target)
	if err != nil {
		http.Error(w, "invalid target URL", http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = r.URL.Path
		req.URL.RawQuery = r.URL.RawQuery
		req.Host = targetURL.Host
	}

	proxy.ErrorHandler = func(writer http.ResponseWriter, request *http.Request, proxyError error) {
		log.Printf("proxy request failed: %v", proxyError)
		http.Error(writer, "upstream service unavailable", http.StatusBadGateway)
	}

	log.Printf("Routing %s %s to %s", r.Method, r.URL.RequestURI(), target)
	proxy.ServeHTTP(w, r)
}
