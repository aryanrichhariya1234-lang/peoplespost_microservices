package handlers

import (
"context"
"crypto/sha256"
"encoding/hex"
"encoding/json"
"net/http"
"os"
"time"


"auth-service/internal/config"
"auth-service/pkg/utils"
"go.mongodb.org/mongo-driver/mongo/options"
"github.com/golang-jwt/jwt/v5"
"go.mongodb.org/mongo-driver/bson"
"go.mongodb.org/mongo-driver/bson/primitive"
"golang.org/x/crypto/bcrypt"


)

type signupRequest struct {
Name            string `json:"name"`
Email           string `json:"email"`
Password        string `json:"password"`
PasswordConfirm string `json:"passwordConfirm"`
GovernmentID    string `json:"governmentId"`
}

type loginRequest struct {
Email    string `json:"email"`
Password string `json:"password"`
}

// ================= SIGNUP =================
func SignUp(w http.ResponseWriter, r *http.Request) {
var body signupRequest


if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
	utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
		"status": "fail", "message": "Invalid request body",
	})
	return
}

if body.Name == "" || body.Email == "" || body.Password == "" {
	utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
		"status": "fail", "message": "Name, email and password required",
	})
	return
}

if body.Password != body.PasswordConfirm {
	utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
		"status": "fail", "message": "Passwords do not match",
	})
	return
}

role := "citizen"
if body.GovernmentID != "" {
	role = "official"
}

hashedPassword, err := bcrypt.GenerateFromPassword([]byte(body.Password), 12)
if err != nil {
	utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
		"status": "error", "message": "Password hashing failed",
	})
	return
}

user := bson.M{
	"name":         body.Name,
	"email":        body.Email,
	"password":     string(hashedPassword),
	"role":         role,
	"governmentId": body.GovernmentID,
	"createdAt":    time.Now(),
}

res, err := config.DB.Collection("users").InsertOne(context.Background(), user)
if err != nil {
	utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
		"status": "error", "message": "User creation failed",
	})
	return
}

id := res.InsertedID.(primitive.ObjectID)
token := generateToken(id.Hex())

setTokenCookie(w, token)

utils.JSON(w, http.StatusOK, map[string]interface{}{
	"status": "success",
	"token":  token,
	"user": map[string]interface{}{
		"id":    id.Hex(),
		"name":  body.Name,
		"email": body.Email,
		"role":  role,
	},
})


}

// ================= LOGIN =================
func Login(w http.ResponseWriter, r *http.Request) {
var body loginRequest


if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
	utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
		"status": "fail", "message": "Invalid request body",
	})
	return
}

if body.Email == "" || body.Password == "" {
	utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
		"status": "fail", "message": "Email and password required",
	})
	return
}

var user bson.M
err := config.DB.Collection("users").
	FindOne(context.Background(), bson.M{"email": body.Email}).
	Decode(&user)

if err != nil {
	utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
		"status": "fail", "message": "Invalid credentials",
	})
	return
}

passwordHash, ok := user["password"].(string)
if !ok {
	utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
		"status": "error", "message": "Invalid user data",
	})
	return
}

if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(body.Password)); err != nil {
	utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
		"status": "fail", "message": "Invalid credentials",
	})
	return
}

id := user["_id"].(primitive.ObjectID)
token := generateToken(id.Hex())

setTokenCookie(w, token)

utils.JSON(w, http.StatusOK, map[string]interface{}{
	"status": "success",
	"token":  token,
})


}

// ================= UPDATE PASSWORD =================
func UpdatePassword(w http.ResponseWriter, r *http.Request) {
val := r.Context().Value("userID")
userID, ok := val.(string)
if !ok {
utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
"status": "fail", "message": "Unauthorized",
})
return
}


objID, err := primitive.ObjectIDFromHex(userID)
if err != nil {
	utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
		"status": "fail", "message": "Invalid user ID",
	})
	return
}

var body struct {
	CurrentPassword string `json:"currentPassword"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"passwordConfirm"`
}

if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
	utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
		"status": "fail", "message": "Invalid body",
	})
	return
}

var user bson.M
err = config.DB.Collection("users").
	FindOne(context.Background(), bson.M{"_id": objID}).
	Decode(&user)

if err != nil {
	utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
		"status": "fail", "message": "User not found",
	})
	return
}

passwordHash, ok := user["password"].(string)
if !ok {
	utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
		"status": "error", "message": "Invalid user data",
	})
	return
}

if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(body.CurrentPassword)); err != nil {
	utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
		"status": "fail", "message": "Incorrect current password",
	})
	return
}

if body.Password != body.PasswordConfirm {
	utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
		"status": "fail", "message": "Passwords do not match",
	})
	return
}

hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(body.Password), 12)

_, err = config.DB.Collection("users").UpdateOne(
	context.Background(),
	bson.M{"_id": objID},
	bson.M{"$set": bson.M{"password": string(hashedPassword)}},
)

if err != nil {
	utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
		"status": "error", "message": "Update failed",
	})
	return
}

token := generateToken(userID)
setTokenCookie(w, token)

utils.JSON(w, http.StatusOK, map[string]interface{}{
	"status": "success", "message": "Password updated",
})


}

// ================= LOGOUT =================
func Logout(w http.ResponseWriter, r *http.Request) {
http.SetCookie(w, &http.Cookie{
Name:     "token",
Value:    "",
Path:     "/",
HttpOnly: true,
Secure:   false,
Expires:  time.Unix(0, 0),
MaxAge:   -1,
})


utils.JSON(w, http.StatusOK, map[string]interface{}{
	"status": "success", "message": "Logged out",
})


}

// ================= FORGOT PASSWORD =================
func ForgotPassword(w http.ResponseWriter, r *http.Request) {
var body struct {
Email string `json:"email"`
}


if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
	utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
		"status": "fail", "message": "Invalid body",
	})
	return
}

var user bson.M
err := config.DB.Collection("users").
	FindOne(context.Background(), bson.M{"email": body.Email}).
	Decode(&user)

if err != nil {
	utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
		"status": "fail", "message": "User not found",
	})
	return
}

tokenData, _ := utils.GenerateResetToken()
objID := user["_id"].(primitive.ObjectID)

_, _ = config.DB.Collection("users").UpdateOne(
	context.Background(),
	bson.M{"_id": objID},
	bson.M{"$set": bson.M{
		"passwordResetToken": tokenData.HashedToken,
		"passwordResetTime":  tokenData.ExpiresAt,
	}},
)

utils.JSON(w, http.StatusOK, map[string]interface{}{
	"status": "success", "message": "Reset token generated",
})


}

// ================= RESET PASSWORD =================
func ResetPassword(w http.ResponseWriter, r *http.Request) {
token := r.URL.Query().Get("token")


var body struct {
	Password string `json:"password"`
}

if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
	utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
		"status": "fail", "message": "Invalid body",
	})
	return
}

hash := sha256.Sum256([]byte(token))
hashedToken := hex.EncodeToString(hash[:])

var user bson.M
err := config.DB.Collection("users").FindOne(context.Background(), bson.M{
	"passwordResetToken": hashedToken,
	"passwordResetTime":  bson.M{"$gt": time.Now()},
}).Decode(&user)

if err != nil {
	utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
		"status": "fail", "message": "Invalid or expired token",
	})
	return
}

hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(body.Password), 12)
objID := user["_id"].(primitive.ObjectID)

_, _ = config.DB.Collection("users").UpdateOne(
	context.Background(),
	bson.M{"_id": objID},
	bson.M{"$set": bson.M{
		"password":           string(hashedPassword),
		"passwordResetToken": "",
	}},
)

utils.JSON(w, http.StatusOK, map[string]interface{}{
	"status": "success",
	"token":  generateToken(objID.Hex()),
})


}

// ================= HELPERS =================
func generateToken(id string) string {
token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
"id":  id,
"exp": time.Now().Add(10 * 24 * time.Hour).Unix(),
})


tokenStr, _ := token.SignedString([]byte(os.Getenv("SECRET")))
return tokenStr


}

func setTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		HttpOnly: true,
		Secure:   false, // 🔥 for now
		// SameSite: http.SameSiteNoneMode,
		
	
		Path:     "/",
	})
}




// ================= GET ME =================
func GetMe(w http.ResponseWriter, r *http.Request) {
val := r.Context().Value("userID")
userID, ok := val.(string)
if !ok {
utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
"status": "fail", "message": "Unauthorized",
})
return
}


objID, err := primitive.ObjectIDFromHex(userID)
if err != nil {
	utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
		"status": "fail", "message": "Invalid user ID",
	})
	return
}

userCollection := config.DB.Collection("users")
postCollection := config.DB.Collection("posts")

var user bson.M
if err := userCollection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&user); err != nil {
	utils.JSON(w, http.StatusNotFound, map[string]interface{}{
		"status": "fail", "message": "User not found",
	})
	return
}

cursor, err := postCollection.Find(context.Background(), bson.M{"user": objID})
if err != nil {
	utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
		"status": "error", "message": "Failed to fetch posts",
	})
	return
}
defer cursor.Close(context.Background())

var posts []bson.M
if err := cursor.All(context.Background(), &posts); err != nil {
	utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
		"status": "error", "message": "Failed to parse posts",
	})
	return
}

utils.JSON(w, http.StatusOK, map[string]interface{}{
	"status": "success",
	"user":   user,
	"data":   posts,
})


}

// ================= UPDATE ME =================
func UpdateMe(w http.ResponseWriter, r *http.Request) {
val := r.Context().Value("userID")
userID, ok := val.(string)
if !ok {
utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
"status": "fail", "message": "Unauthorized",
})
return
}


objID, err := primitive.ObjectIDFromHex(userID)
if err != nil {
	utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
		"status": "fail", "message": "Invalid user ID",
	})
	return
}

var body map[string]interface{}
if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
	utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
		"status": "fail", "message": "Invalid body",
	})
	return
}

updateFields := bson.M{}

if name, ok := body["name"].(string); ok && name != "" {
	updateFields["name"] = name
}
if email, ok := body["email"].(string); ok && email != "" {
	updateFields["email"] = email
}

if len(updateFields) == 0 {
	utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
		"status": "fail", "message": "No valid fields to update",
	})
	return
}

collection := config.DB.Collection("users")

var updated bson.M
err = collection.FindOneAndUpdate(
	context.Background(),
	bson.M{"_id": objID},
	bson.M{"$set": updateFields},
	options.FindOneAndUpdate().SetReturnDocument(options.After),
).Decode(&updated)

if err != nil {
	utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
		"status": "error", "message": "Update failed",
	})
	return
}

utils.JSON(w, http.StatusOK, map[string]interface{}{
	"status": "success",
	"user":   updated,
})


}

// ================= DELETE ME =================
func DeleteMe(w http.ResponseWriter, r *http.Request) {
val := r.Context().Value("userID")
userID, ok := val.(string)
if !ok {
utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
"status": "fail", "message": "Unauthorized",
})
return
}


objID, err := primitive.ObjectIDFromHex(userID)
if err != nil {
	utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
		"status": "fail", "message": "Invalid user ID",
	})
	return
}

collection := config.DB.Collection("users")

res, err := collection.DeleteOne(context.Background(), bson.M{"_id": objID})
if err != nil || res.DeletedCount == 0 {
	utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
		"status": "error", "message": "Delete failed",
	})
	return
}

utils.JSON(w, http.StatusOK, map[string]interface{}{
	"status": "success",
})


}
