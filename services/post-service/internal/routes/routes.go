package routes

import (
	"net/http"
	"strings"

	"post-service/internal/handlers"
	"post-service/internal/middleware"
)

func SetupRouter() http.Handler {
	mux := http.NewServeMux()

	// ================= ROOT =================
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"success","message":"API running"}`))
	})

	
	// ================= POSTS (RESTFUL) =================

	// GET all posts + CREATE post
	mux.HandleFunc("/api/v1/posts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handlers.GetAllPosts(w, r)

		case http.MethodPost:
			middleware.Protect(handlers.CreatePost)(w, r)

		default:
			http.NotFound(w, r)
		}
	})

	// UPDATE / DELETE / LIKE (with ID)
	mux.HandleFunc("/api/v1/posts/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/posts/")

		// /:id/like
		if strings.HasSuffix(path, "/like") && r.Method == http.MethodPost {
			middleware.Protect(handlers.ToggleLike)(w, r)
			return
		}

		// /:id
		switch r.Method {
		case http.MethodPatch:
			middleware.Protect(handlers.UpdatePost)(w, r)

		case http.MethodDelete:
			middleware.Protect(handlers.DeletePost)(w, r)

		default:
			http.NotFound(w, r)
		}
	})

	// ================= CORS =================
	return mux
}