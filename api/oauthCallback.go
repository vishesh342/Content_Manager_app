package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

const (
	authorization = "authorization_code"
	clientID     = "86zk1jyrzrqnfw"
	clientSecret = "WPL_AP1.xalSibXXFTXvniGr.0Op+Zw=="
	redirectURI  = "http://localhost:8080/oauth/linkedin/callback"
	tokenURL = "https://www.linkedin.com/oauth/v2/accessToken"
)

type linkedinOAuthRequest struct {
    Code  string `uri:"code" binding:"required"`
    State string `uri:"state" binding:"required"`
}

type accessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}
type successResponse struct{
	Remark string `json:"remark"`
}

// Handle LinkedIn OAuth callback
func(server *Server) handleLinkedInCallback(ctx *gin.Context) {
	// Get 'code' and 'state' from query parameters
	var req linkedinOAuthRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Exchange authorization code for access token
	_, err := getLinkedInAccessToken(req.Code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get access token"})
		return
	}
	resp := successResponse{
		Remark: "Request Authorised...",
	}
	// Store token in database (for now, just print it)
	// TODO: Store token securely in the database
	ctx.JSON(http.StatusOK, resp)
}

// Exchange authorization code for access token
func getLinkedInAccessToken(code string) (string, error) {

	// Prepare request body
	data := url.Values{}
	data.Set("grant_type", authorization)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)

	// Make POST request to get access token
	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse JSON response
	var tokenResponse accessTokenResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", err
	}

	return tokenResponse.AccessToken, nil
}