package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

// ================= PROXY =================
func newProxy(target string) *httputil.ReverseProxy {
	u, _ := url.Parse(target)
	return httputil.NewSingleHostReverseProxy(u)
}

// ================= CORS =================
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")

		if r.Method == http.MethodOptions {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	authURL := os.Getenv("AUTH_SERVICE_URL")
	postURL := os.Getenv("POST_SERVICE_URL")
	aiURL := os.Getenv("AI_SERVICE_URL")

	authProxy := newProxy(authURL)
	postProxy := newProxy(postURL)
	aiProxy := newProxy(aiURL)

	mux := http.NewServeMux()

	// routes
	mux.Handle("/api/v1/auth", authProxy)
	mux.Handle("/api/v1/posts", postProxy)
	mux.Handle("/api/v1/posts/", postProxy)
	mux.Handle("/api/v1/ai", aiProxy)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"gateway running"}`))
	})

	log.Println("Gateway running on port 4000")
	http.ListenAndServe(":4000", cors(mux))
}