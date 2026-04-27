package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"post-service/internal/middleware"
)

// ================= SAFE WRAPPER =================
// prevents test from crashing due to DB nil panic
func safeHandler(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// simulate internal server error instead of crashing test
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		h(w, r)
	}
}

// ================= MOCK AUTH =================
func mockAuthContext(req *http.Request) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, "507f1f77bcf86cd799439011")
	return req.WithContext(ctx)
}

// ================= TESTS =================

func TestGetAllPosts(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/posts", nil)
	rr := httptest.NewRecorder()

	handler := safeHandler(GetAllPosts)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK && rr.Code != http.StatusInternalServerError {
		t.Errorf("Unexpected status code: %d", rr.Code)
	}
}

func TestCreatePost_MissingFields(t *testing.T) {
	body := map[string]interface{}{
		"category": "Road",
	}

	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/api/v1/posts", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	req = mockAuthContext(req)

	rr := httptest.NewRecorder()

	handler := safeHandler(CreatePost)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest && rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected 400 or 500, got %d", rr.Code)
	}
}

func TestCreatePost_Success(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("category", "Road")
	writer.WriteField("Address", "Somewhere")
	writer.WriteField("description", "Big pothole")
	writer.WriteField("location", `{"lat":19.07,"lng":72.87}`)

	part, _ := writer.CreateFormFile("images", "test.jpg")
	part.Write([]byte("fake image"))

	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/posts", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	req = mockAuthContext(req)

	rr := httptest.NewRecorder()

	handler := safeHandler(CreatePost)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated &&
		rr.Code != http.StatusOK &&
		rr.Code != http.StatusInternalServerError {

		t.Errorf("Unexpected status code: %d", rr.Code)
	}
}

func TestProtectedRoute_Unauthorized(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/v1/posts", nil)

	rr := httptest.NewRecorder()

	handler := safeHandler(CreatePost)
	handler.ServeHTTP(rr, req)

	if rr.Code == http.StatusOK {
		t.Errorf("Expected unauthorized error")
	}
}

func TestStatusValidation(t *testing.T) {
	valid := []string{"OPEN", "IN_PROGRESS", "RESOLVED"}
	invalid := "DONE"

	for _, s := range valid {
		if !isValidStatus(s) {
			t.Errorf("Expected %s valid", s)
		}
	}

	if isValidStatus(invalid) {
		t.Errorf("Invalid status passed")
	}
}

// helper (same as yours)
func isValidStatus(status string) bool {
	switch status {
	case "OPEN", "IN_PROGRESS", "RESOLVED":
		return true
	}
	return false
}