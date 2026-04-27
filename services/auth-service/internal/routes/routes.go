package routes

import (
	"net/http"


	"auth-service/internal/handlers"
	"auth-service/internal/middleware"
)

func SetupRouter() http.Handler {
	mux := http.NewServeMux()

	// ================= ROOT =================
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"success","message":"API running"}`))
	})

	// ================= USERS =================
	mux.HandleFunc("/api/v1/users/me", middleware.Protect(handlers.GetMe))

	mux.HandleFunc("/api/v1/users/forgotPassword", handlers.ForgotPassword)
	mux.HandleFunc("/api/v1/users/updateMe", middleware.Protect(handlers.UpdateMe))
	mux.HandleFunc("/api/v1/users/updatePassword", middleware.Protect(handlers.UpdatePassword))
	mux.HandleFunc("/api/v1/users/signup", handlers.SignUp)
	mux.HandleFunc("/api/v1/users/logout", handlers.Logout)
	mux.HandleFunc("/api/v1/users/login", handlers.Login)
	mux.HandleFunc("/api/v1/users/deleteMe", middleware.Protect(handlers.DeleteMe))
	mux.HandleFunc("/api/v1/users/resetPassword", handlers.ResetPassword)

	
	
	return mux
}