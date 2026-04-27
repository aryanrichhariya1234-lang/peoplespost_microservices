package main

import (
	"api-gateway/internal/config"
	"api-gateway/internal/middleware"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)


// ================= COOKIE HELPERS =================

// remove specific attribute from Set-Cookie
func removeAttr(cookie, attr string) string {
	parts := strings.Split(cookie, ";")
	var result []string

	for _, p := range parts {
		if !strings.HasPrefix(strings.TrimSpace(strings.ToLower(p)), strings.ToLower(attr)) {
			result = append(result, p)
		}
	}

	return strings.Join(result, ";")
}


// ================= PROXY =================

func newProxy(target string) *httputil.ReverseProxy {
	u, err := url.Parse(target)
	if err != nil {
		log.Fatal("Invalid proxy target:", target)
	}

	proxy := httputil.NewSingleHostReverseProxy(u)

	// 🔥 Fix cookies here
	proxy.ModifyResponse = func(res *http.Response) error {
		cookies := res.Header["Set-Cookie"]
	
		for i, c := range cookies {
	
			// remove domain only in dev
			if os.Getenv("ENV") == "development" {
				c = removeAttr(c, "Domain")
				c = removeAttr(c, "Secure")
			}
	
			// set proper SameSite
			if !strings.Contains(strings.ToLower(c), "samesite") {
				if os.Getenv("ENV") == "production" {
					c += "; SameSite=None; Secure"
				} else {
					c += "; SameSite=Lax"
				}
			}
	
			cookies[i] = c
		}
	
		if len(cookies) > 0 {
			res.Header["Set-Cookie"] = cookies
		}
	
		return nil
	}

	// optional: better error logging
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Println("Proxy error:", err)
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Bad Gateway"))
	}

	return proxy
}


// ================= MAIN =================

func main() {
	config.LoadEnv()

	authURL := os.Getenv("AUTH_SERVICE_URL")
	postURL := os.Getenv("POST_SERVICE_URL")
	aiURL := os.Getenv("AI_SERVICE_URL")

	authProxy := newProxy(authURL)
	postProxy := newProxy(postURL)
	aiProxy := newProxy(aiURL)

	mux := http.NewServeMux()

	// ✅ IMPORTANT: use trailing slash for proper routing
	mux.Handle("/api/v1/users/", authProxy)
	
	mux.Handle("/api/v1/posts", postProxy)
	mux.Handle("/api/v1/ai/", aiProxy)

	// health route
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"gateway running"}`))
	})

	port := config.PORT
	if port == "" {
		port = "4000"
	}

	log.Println("Gateway running on port", port)

	err := http.ListenAndServe(":"+port, middleware.CORS(mux))
	if err != nil {
		log.Fatal("Server failed:", err)
	}
}