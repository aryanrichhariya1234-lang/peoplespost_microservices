package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"api-gateway/internal/config"
	"api-gateway/internal/middleware"
)

// ================= PROXY =================
func newProxy(target string) *httputil.ReverseProxy {
	u, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Invalid service URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(u)

	// 🔥 IMPORTANT: remove downstream CORS headers
	proxy.ModifyResponse = func(resp *http.Response) error {
		resp.Header.Del("Access-Control-Allow-Origin")
		resp.Header.Del("Access-Control-Allow-Credentials")
		resp.Header.Del("Access-Control-Allow-Headers")
		resp.Header.Del("Access-Control-Allow-Methods")
		return nil
	}

	return proxy
}

func main() {
	config.LoadEnv()

	authURL := os.Getenv("AUTH_SERVICE_URL")
	postURL := os.Getenv("POST_SERVICE_URL")
	aiURL := os.Getenv("AI_SERVICE_URL")

	authProxy := newProxy(authURL)
	postProxy := newProxy(postURL)
	aiProxy := newProxy(aiURL)

	mux := http.NewServeMux()

	mux.Handle("/api/v1/users/", authProxy)
	mux.Handle("/api/v1/posts", postProxy)
	mux.Handle("/api/v1/posts/", postProxy)
	mux.Handle("/api/v1/ai/", aiProxy)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"gateway running"}`))
	})

	log.Println("Gateway running on port 4000")

	// ✅ ONLY THIS CORS
	log.Fatal(http.ListenAndServe(":4000", middleware.CORS(mux)))
}