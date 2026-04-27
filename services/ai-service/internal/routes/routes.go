package routes

import (
	"net/http"


	"ai-service/internal/handlers"
	"ai-service/internal/middleware"
)

func SetupRouter() http.Handler {
	mux := http.NewServeMux()

	// ================= ROOT =================
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"success","message":"API running"}`))
	})

	mux.HandleFunc("/api/v1/ai/insights", middleware.Protect(handlers.GetDashboardInsights))





	// ================= CORS =================
	return mux
}