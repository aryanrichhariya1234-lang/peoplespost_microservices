package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"ai-service/internal/config"
	"ai-service/pkg/utils"
)

// ================= DASHBOARD INSIGHTS =================
func GetDashboardInsights(w http.ResponseWriter, r *http.Request) {
	config.LoadEnv()

	// ✅ 2. Call Post Service (analytics endpoint)
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(config.POST_SERVICE_URL + "/api/v1/posts/analytics")

	if err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Failed to fetch analytics from post service",
		})
		return
	}
	defer resp.Body.Close()

	var analyticsResp map[string]interface{}

	if err := json.NewDecoder(resp.Body).Decode(&analyticsResp); err != nil {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Invalid analytics response",
		})
		return
	}

	data, ok := analyticsResp["data"].(map[string]interface{})
	if !ok {
		utils.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"message": "Malformed analytics data",
		})
		return
	}

	// ✅ 3. Extract data safely
	cityStats := data["cityStats"]
	categoryStats := data["categoryStats"]
	userStats := data["userStats"]

	currentWeek := int(data["currentWeek"].(float64))
	previousWeek := int(data["previousWeek"].(float64))
	totalReports := int(data["totalReports"].(float64))

	// ✅ 4. Compute trend
	trend := 100.0
	if previousWeek != 0 {
		trend = float64(currentWeek-previousWeek) / float64(previousWeek) * 100
	}

	estimatedImpact := totalReports * 50

	priority := "LOW"
	if trend > 30 || totalReports > 20 {
		priority = "HIGH"
	} else if trend > 10 {
		priority = "MEDIUM"
	}

	// ✅ 5. Build prompt
	cityJSON, _ := json.Marshal(cityStats)
	categoryJSON, _ := json.Marshal(categoryStats)
	userJSON, _ := json.Marshal(userStats)

	prompt := fmt.Sprintf(`
You are a smart city decision intelligence system.

Give:
1. Most affected city
2. Most common issue
3. Most active reporter
4. Trend (with %%)
5. Priority level
6. Estimated people affected
7. Action recommendation

City Stats: %s
Category Stats: %s
User Stats: %s

Trend: %.2f%%
Total Reports: %d
Impact: %d
Priority: %s
`, cityJSON, categoryJSON, userJSON, trend, totalReports, estimatedImpact, priority)

	// ✅ 6. Call Gemini
	insights := "AI unavailable"

	responseText, err := callAI(prompt)
	if err == nil && responseText != "" {
		insights = responseText
	}
	if err != nil {
		fmt.Println(err)
	}

	// ✅ 7. Final response
	result := map[string]interface{}{
		"status":   "success",
		"insights": insights,
		"meta": map[string]interface{}{
			"totalReports":    totalReports,
			"trend":           trend,
			"priority":        priority,
			"estimatedImpact": estimatedImpact,
		},
	}

	utils.JSON(w, http.StatusOK, result)
}

// ================= GEMINI CALL =================
func callAI(prompt string) (string, error) {
	apiKey := os.Getenv("OPENROUTER_API_KEY")

	if apiKey == "" {
		return "", fmt.Errorf("missing OPENROUTER_API_KEY")
	}

	reqBody := map[string]interface{}{
		"model": "openai/gpt-4o-mini",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a smart city decision intelligence system. Analyze city reports and provide concise actionable insights.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.3,
		"max_tokens":  800,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(
		"POST",
		"https://openrouter.ai/api/v1/chat/completions",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// Optional but recommended
	req.Header.Set("HTTP-Referer", "http://localhost")
	req.Header.Set("X-Title", "Smart City AI Service")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	// Debug if needed
	if errObj, ok := result["error"]; ok {
		b, _ := json.Marshal(errObj)
		return "", fmt.Errorf("openrouter error: %s", string(b))
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		b, _ := json.Marshal(result)
		return "", fmt.Errorf("no choices returned: %s", string(b))
	}

	choice := choices[0].(map[string]interface{})
	message := choice["message"].(map[string]interface{})

	content, ok := message["content"].(string)
	if !ok {
		return "", fmt.Errorf("invalid response format")
	}

	return content, nil
}
