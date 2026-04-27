package middleware

import (
"context"
"net/http"
"os"
"strings"

"github.com/golang-jwt/jwt/v5"


)

// optional: define a custom type to avoid context key collisions
type contextKey string

const userIDKey contextKey = "userID"

func Protect(next http.HandlerFunc) http.HandlerFunc {
return func(w http.ResponseWriter, r *http.Request) {
var tokenStr string


	// 1. Try cookie
	cookie, err := r.Cookie("token")
	if err == nil {
		tokenStr = cookie.Value
	}

	// 2. Try Authorization header
	if tokenStr == "" {
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			tokenStr = strings.TrimPrefix(auth, "Bearer ")
		}
	}

	// 3. No token
	if tokenStr == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 4. Parse & validate token
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		// Optional: enforce signing method
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, http.ErrAbortHandler
		}
		return []byte(os.Getenv("SECRET")), nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// 5. Extract claims safely
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	userID, ok := claims["id"].(string)
	if !ok || userID == "" {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	// 6. Add to context
	ctx := context.WithValue(r.Context(), userIDKey, userID)

	// 7. Continue
	next(w, r.WithContext(ctx))
}


}

// Helper to get userID safely in handlers
func GetUserID(r *http.Request) (string, bool) {
id, ok := r.Context().Value(userIDKey).(string)
return id, ok
}
