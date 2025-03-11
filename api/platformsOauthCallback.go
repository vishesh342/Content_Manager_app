package api

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/vishesh342/content-manager/db/sqlc"
	token "github.com/vishesh342/content-manager/tokens"
)

const (
	authorization = "authorization_code"
	clientID     = "86zk1jyrzrqnfw"
	clientSecret = "WPL_AP1.xalSibXXFTXvniGr.0Op+Zw=="
	redirectURI  = "http://localhost:8080/oauth/linkedin/callback"
	tokenURL = "https://www.linkedin.com/oauth/v2/accessToken"
	userIdURL = "https://api.linkedin.com/v2/me"
)

type linkedinOAuthRequest struct {
    Code  string `uri:"code" binding:"required"`
    State string `uri:"state" binding:"required"`
}

type accessTokenResponse struct {
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
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
	tokenResp , err := getLinkedInAccessToken(req.Code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get access token"})
		return
	}
	// Store token in database (for now, just print it)
	// TODO: Store token securely in the database
	authPayload := ctx.MustGet(authorizationPayload).(*token.Payload)

	err = server.createSocialAccount(ctx, authPayload,tokenResp)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp := successResponse{
		Remark: "Request Authorised",
	}
	ctx.JSON(http.StatusOK, resp)
}

// Exchange authorization code for access token
func getLinkedInAccessToken(code string) (accessTokenResponse, error) {
	var tokenResponse accessTokenResponse

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
		return tokenResponse, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return tokenResponse, err
	}

	// Parse JSON response
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return tokenResponse, err
	}

	return tokenResponse, nil
}
type LinkedInUser struct {
	ID                 string `json:"id"`
}
func (server *Server) createSocialAccount(ctx *gin.Context,authPayload *token.Payload, tokenResp accessTokenResponse) error {
	var user LinkedInUser
	req, _ := http.NewRequest("GET", userIdURL, nil)
	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return  err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return  err
	}
	if err := json.Unmarshal(body, &user); err != nil {
		return  err
	}

	// Get Platform-ID
	platform,err := server.connector.GetPlatform(ctx,"LinkedIn")
	if err != nil{
		return err
	}	
	// Check krna h ki agar user ka platform created hai -> access token update kr do nahi toh account create kr do. 
	arg := db.GetAccountParams{
		Username: authPayload.Username,
		PlatformID: platform.ID,
	}
	_, err = server.connector.GetAccount(ctx,arg)
	if err != nil {
		// User not found
		if err == sql.ErrNoRows {
			arg := db.CreateAccountParams{
				Username: authPayload.Username,
				PlatformID: platform.ID,
				PlatformUsername: pgtype.Text{String: user.ID},
				AccessToken: tokenResp.AccessToken,
				RefreshToken: pgtype.Text{String: tokenResp.RefreshToken},
				ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Second*time.Duration(tokenResp.ExpiresIn)), Valid: true},
			}
			_, err = server.connector.CreateAccount(ctx,arg)
			if err != nil{
				return err
			}
		}
	}
	updateArg:= db.UpdateAccountParams {
		Username: authPayload.Username,
		PlatformID: platform.ID,
		AccessToken: tokenResp.AccessToken,
		RefreshToken: pgtype.Text{String: tokenResp.RefreshToken},
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Second*time.Duration(tokenResp.ExpiresIn)), Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: time.Now(),Valid: true},
	}
	err = server.connector.UpdateAccount(ctx,updateArg)
	if err != nil{
		return err
	}
	return nil
}