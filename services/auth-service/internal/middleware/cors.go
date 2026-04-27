package middleware

import "net/http"

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		origin := r.Header.Get("Origin")

		// ✅ Allowed origins
		allowedOrigins := map[string]bool{
			"http://localhost:3000":          true,
			"http://127.0.0.1:3000":          true,
			"http://localhost:3001":          true,
			"https://peoplespost.vercel.app": true,
		}

		// 🔥 Important for caching proxies
		w.Header().Set("Vary", "Origin")

		// ✅ Set origin only if allowed
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// ✅ Allowed headers
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// ✅ Allowed methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")

		// ✅ Expose headers (important for cookies/debugging)
		w.Header().Set("Access-Control-Expose-Headers", "Set-Cookie")

		// ✅ Handle preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}