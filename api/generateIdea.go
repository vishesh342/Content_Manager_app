package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	apiKey       = "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" // Replace with your actual API key
	systemPrompt = "For every category the user mentions, generate exactly %s trending topics that emerging in the past %s days.Ensure the topics are fresh, specific, and not generic. Avoid outdated or evergreen content."
	userPrompt   = "Fetch the latest and most relevant trending topic ideas from each the following domains: %s. The news should: - Be from reputable sources - Include articles with high engagement - Be specific, fresh, and not generic. Format the response strictly as a JSON object with the following keys for each news item: - title: A short, engaging headline summarizing the news - summary: A concise 3–5 sentence digest covering key insights - significance: A 1–2 sentence explanation of why it matters - links: An object with source and url. Note: Return only the JSON object. Do not include any additional text, markdown, or explanation."
	model        = "mistralai/mistral-7b-instruct:free" // Replace with the model you are using
	requestUrl   = "https://openrouter.ai/api/v1/chat/completions" // Replace with the actual URL
)

// Define the message structure
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenRouterRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Response_Format map[string]interface{} `json:"response_format,omitempty"`
	Temperature float64   `json:"temperature"`
}

// Define the response structure
type OpenRouterResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}
type Payload struct {
	Day      string   `json:"day"`
	Count    string   `json:"count"`
	Category []string `json:"category"`
}
var openRouterResponseSchema = map[string]interface{}{
    "type": "json_schema",
    "json_schema": map[string]interface{}{
        "type": "object",
        "patternProperties": map[string]interface{}{
            "^.*$": map[string]interface{}{
                "type": "array",
                "items": map[string]string{"$ref": "#/definitions/newsItem"},
            },
        },
        "additionalProperties": false,
        "definitions": map[string]interface{}{
            "newsItem": map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "title": map[string]string{"type": "string"},
                    "summary": map[string]string{"type": "string"},
                    "significance": map[string]string{"type": "string"},
                    "links": map[string]interface{}{
                        "type": "object",
                        "properties": map[string]interface{}{
                            "source": map[string]string{"type": "string"},
                            "url": map[string]string{
                                "type": "string",
                                "format": "uri",
                            },
                        },
                        "required": []string{"source", "url"},
                        "additionalProperties": false,
                    },
                },
                "required": []string{"title", "summary", "significance", "links"},
                "additionalProperties": false,
            },
        },
    },
}

func(server *Server) generateIdeaHandler(c *gin.Context) {
	var request Payload
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	response, err := callOpenRouter(apiKey, request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to contact OpenRouter"})
		return
	}

	// Send response back to client
	c.JSON(http.StatusOK, response)

}

func callOpenRouter(apiKey string, request Payload) (string, error) {
	catergory := strings.Join(request.Category, ", ")
	user := fmt.Sprintf(userPrompt, catergory)
	system := fmt.Sprintf(systemPrompt, request.Count, request.Day)

	requestData := OpenRouterRequest{
		Model: model,
		Messages: []Message{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Response_Format: openRouterResponseSchema,
		Temperature: 0.8,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request data: %v", err)
	}

	req, err := http.NewRequest("POST", requestUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	var response OpenRouterResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if len(response.Choices) > 0 {
		return response.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no content in response")
}