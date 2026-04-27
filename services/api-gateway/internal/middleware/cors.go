package middleware

import (
	"log"
	"net/http"
)

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		origin := r.Header.Get("Origin")

		allowedOrigins := map[string]bool{
			// 🔧 Local dev
			"http://localhost:3000":  true,
			"http://127.0.0.1:3000": true,

			// 🔥 Production (UPDATE THESE)
			"https://peoplespost.vercel.app/": true,
			"http://100.30.218.48:3000":  true,
			"https://app.yoursite.com":   true,
		}

		w.Header().Set("Vary", "Origin")

		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		} else if origin != "" {
			log.Println("❌ CORS blocked origin:", origin)
		}

		w.Header().Set("Access-Control-Allow-Headers",
			"Content-Type, Authorization, X-Requested-With, Accept")

		w.Header().Set("Access-Control-Allow-Methods",
			"GET, POST, PATCH, DELETE, OPTIONS")

		w.Header().Set("Access-Control-Expose-Headers", "Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}