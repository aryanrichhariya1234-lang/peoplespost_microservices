package handlers

import (
"context"
"encoding/json"
"net/http"
"strings"
"time"

"fmt"
"mime/multipart"

"github.com/cloudinary/cloudinary-go/v2/api/uploader"
"go.mongodb.org/mongo-driver/bson"
"go.mongodb.org/mongo-driver/bson/primitive"
"go.mongodb.org/mongo-driver/mongo/options"

"post-service/internal/cache"
"post-service/internal/config"
"post-service/internal/middleware"
"post-service/pkg/utils"


)
type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}
type createPostRequest struct {
Category    string      `json:"category"`
Description string      `json:"description"`
Location    interface{} `json:"location"`
Images      []string    `json:"images"`
Status      string      `json:"status,omitempty"`
}
func getCityFromLatLng(ctx context.Context, lat, lng float64) (string, error) {
	url := fmt.Sprintf(
		"https://nominatim.openstreetmap.org/reverse?lat=%f&lon=%f&format=json",
		lat, lng,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "peoplepost-app")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data struct {
		Address struct {
			City    string `json:"city"`
			Town    string `json:"town"`
			Village string `json:"village"`
			State   string `json:"state"`
		} `json:"address"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	if data.Address.City != "" {
		return data.Address.City, nil
	}
	if data.Address.Town != "" {
		return data.Address.Town, nil
	}
	if data.Address.Village != "" {
		return data.Address.Village, nil
	}
	if data.Address.State != "" {
		return data.Address.State, nil
	}

	return "Unknown", nil
}
// ================= CREATE =================
func CreatePost(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
			"status": "fail", "message": "Unauthorized",
		})
		return
	}

	if config.Cloudinary == nil || config.DB == nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status": "error",
			"message": "Server not properly initialized",
		})
		return
	}
	

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status": "fail", "message": "Invalid form data",
		})
		return
	}

	category := r.FormValue("category")
	description := r.FormValue("description")
	locationStr := r.FormValue("location")

	var files []*multipart.FileHeader
	if r.MultipartForm != nil {
		files = r.MultipartForm.File["images"]
	}

	if category == "" || description == "" || locationStr == "" || len(files) == 0 {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status": "fail",
			"message": "category, description, location and images are required",
		})
		return
	}

	// ✅ Parse location
	var loc Location
	if err := json.Unmarshal([]byte(locationStr), &loc); err != nil {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status": "fail", "message": "Invalid location format",
		})
		return
	}

	// ✅ Reverse geocode (with timeout)
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	city, err := getCityFromLatLng(ctx, loc.Lat, loc.Lng)
	if err != nil {
		city = "Unknown" // fallback
	}

	// ✅ Upload images
	var imageURLs []string
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			continue
		}

		uploadRes, err := config.Cloudinary.Upload.Upload(
			r.Context(),
			file,
			uploader.UploadParams{Folder: "posts"},
		)
		file.Close()

		if err == nil {
			imageURLs = append(imageURLs, uploadRes.SecureURL)
		}
	}

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status": "fail", "message": "Invalid user ID",
		})
		return
	}

	// ✅ Final post object
	post := bson.M{
		"user":        objID,
		"category":    category,
		"description": description,
		"images":      imageURLs,
		"location": bson.M{
			"lat": loc.Lat,
			"lng": loc.Lng,
		},
		"city":      city, // 🔥 separate field
		"status":    "OPEN",
		"likes":     []interface{}{},
		"createdAt": time.Now(),
	}

	res, err := config.DB.Collection("posts").InsertOne(context.Background(), post)
	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status": "error", "message": "Failed to create post",
		})
		return
	}

	cache.Delete("posts")
	cache.Delete("posts:analytics") // 🔥 updated cache key

	utils.JSON(w, http.StatusCreated, map[string]interface{}{
		"status": "success", "data": res.InsertedID,
	})
}
// ================= GET =================
func GetAllPosts(w http.ResponseWriter, r *http.Request) {
var cached []bson.M

if err := cache.Get("posts", &cached); err == nil && len(cached) > 0 {
	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status": "success", "data": cached,
	})
	return
}

cursor, err := config.DB.Collection("posts").Find(context.Background(), bson.M{})
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

_ = cache.Set("posts", posts, 30*time.Minute)

utils.JSON(w, http.StatusOK, map[string]interface{}{
	"status": "success", "data": posts,
})


}

// ================= UPDATE =================
func UpdatePost(w http.ResponseWriter, r *http.Request) {
	// ✅ Only verify token
	_, ok := middleware.GetUserID(r)
	if !ok {
		utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
			"status": "fail", "message": "Unauthorized",
		})
		return
	}

	id := extractID(r)
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status": "fail", "message": "Invalid ID",
		})
		return
	}

	collection := config.DB.Collection("posts")

	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status": "fail", "message": "Invalid body",
		})
		return
	}

	update := bson.M{}

	if status, ok := body["status"].(string); ok {
		status = strings.ToUpper(status)

		valid := map[string]bool{
			"OPEN": true, "IN_PROCESS": true, "RESOLVED": true,
		}

		if !valid[status] {
			utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
				"status": "fail", "message": "Invalid status",
			})
			return
		}

		update["status"] = status
	}

	if len(update) == 0 {
		utils.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status": "fail", "message": "No valid fields",
		})
		return
	}

	var updated bson.M
	err = collection.FindOneAndUpdate(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{"$set": update},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)

	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status": "error", "message": "Update failed",
		})
		return
	}

	cache.Delete("posts")
	cache.Delete("dashboard:insights")

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status": "success", "data": updated,
	})
}
// ================= DELETE =================
func DeletePost(w http.ResponseWriter, r *http.Request) {
userID, ok := middleware.GetUserID(r)
if !ok {
utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
"msg": "Unauthorized",
})
return
}


id := extractID(r)
objID, _ := primitive.ObjectIDFromHex(id)

collection := config.DB.Collection("posts")

var existing bson.M
if err := collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&existing); err != nil {
	utils.JSON(w, http.StatusNotFound, map[string]interface{}{
		"msg": "Post not found",
	})
	return
}

if existing["user"].(primitive.ObjectID).Hex() != userID {
	utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
		"msg": "Unauthorized",
	})
	return
}

_, err := collection.DeleteOne(context.Background(), bson.M{"_id": objID})
if err != nil {
	utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
		"msg": "Delete failed",
	})
	return
}

cache.Delete("posts")
cache.Delete("dashboard:insights")

utils.JSON(w, http.StatusOK, map[string]interface{}{
	"msg": "Post removed successfully",
})


}

// ================= LIKE =================
func ToggleLike(w http.ResponseWriter, r *http.Request) {
userID, ok := middleware.GetUserID(r)
if !ok {
utils.JSON(w, http.StatusUnauthorized, map[string]interface{}{
"message": "Unauthorized",
})
return
}


id := extractID(r)
objID, _ := primitive.ObjectIDFromHex(id)

collection := config.DB.Collection("posts")

var post bson.M
if err := collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&post); err != nil {
	utils.JSON(w, http.StatusNotFound, map[string]interface{}{
		"message": "Post not found",
	})
	return
}

likes, _ := post["likes"].([]interface{})
already := false
var updated []interface{}

for _, l := range likes {
	uid := l.(bson.M)["user"].(primitive.ObjectID).Hex()
	if uid == userID {
		already = true
		continue
	}
	updated = append(updated, l)
}

if !already {
	uid, _ := primitive.ObjectIDFromHex(userID)
	updated = append(updated, bson.M{"user": uid})
}

_, err := collection.UpdateOne(
	context.Background(),
	bson.M{"_id": objID},
	bson.M{"$set": bson.M{"likes": updated}},
)

if err != nil {
	utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
		"message": "Update failed",
	})
	return
}

cache.Delete("posts")
cache.Delete("dashboard:insights")

utils.JSON(w, http.StatusOK, map[string]interface{}{
	"status":     "success",
	"liked":      !already,
	"likesCount": len(updated),
})


}

// ================= HELPER =================
func extractID(r *http.Request) string {
path := strings.TrimPrefix(r.URL.Path, "/api/v1/posts/")
return strings.Split(path, "/")[0]
}

func GetPostAnalytics(w http.ResponseWriter, r *http.Request) {
	
	ctx := context.Background()
	cacheKey := "posts:analytics"

	// ✅ 1. Try cache first
	var cached map[string]interface{}
	if err := cache.Get(cacheKey, &cached); err == nil {
		utils.JSON(w, http.StatusOK, cached)
		return
	}

	collection := config.DB.Collection("posts")

	// ✅ 2. Aggregations
	cityStats, _ := utils.Aggregate(collection, "$location.city")
	categoryStats, _ := utils.Aggregate(collection, "$category")
	userStats, _ := utils.Aggregate(collection, "$user")

	// ✅ 3. Time-based stats
	now := time.Now()
	last7Days := now.AddDate(0, 0, -7)
	prev7Days := now.AddDate(0, 0, -14)

	currentWeek, _ := collection.CountDocuments(ctx, bson.M{
		"createdAt": bson.M{"$gte": last7Days},
	})

	previousWeek, _ := collection.CountDocuments(ctx, bson.M{
		"createdAt": bson.M{
			"$gte": prev7Days,
			"$lt":  last7Days,
		},
	})

	totalReports, _ := collection.CountDocuments(ctx, bson.M{})

	// ✅ 4. Response
	result := map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"cityStats":     cityStats,
			"categoryStats": categoryStats,
			"userStats":     userStats,
			"totalReports":  totalReports,
			"currentWeek":   currentWeek,
			"previousWeek":  previousWeek,
		},
	}

	// ✅ 5. Cache it
	_ = cache.Set(cacheKey, result, 15*time.Minute)

	utils.JSON(w, http.StatusOK, result)
}