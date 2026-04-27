package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

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

	// 🔥 Remove downstream CORS headers
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

	authProxy := newProxy(config.AUTH_SERVICE_URL)
	postProxy := newProxy(config.POST_SERVICE_URL)
	aiProxy := newProxy(config.AI_SERVICE_URL)

	mux := http.NewServeMux()

	// Routes
	mux.Handle("/api/v1/users/", authProxy)
	mux.Handle("/api/v1/posts", postProxy)
	mux.Handle("/api/v1/posts/", postProxy)
	mux.Handle("/api/v1/ai/", aiProxy)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"gateway running"}`))
	})

	log.Println("Gateway running on port", config.PORT)

	// ✅ Use ENV port
	log.Fatal(http.ListenAndServe(":"+config.PORT, middleware.CORS(mux)))
}