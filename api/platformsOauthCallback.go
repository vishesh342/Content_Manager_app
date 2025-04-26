package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/vishesh342/content-manager/db/sqlc"
	tk "github.com/vishesh342/content-manager/tokens"
)

const (
	authorization = "authorization_code"
	clientID     = "86zk1jyrzrqnfw"
	clientSecret = "WPL_AP1.xalSibXXFTXvniGr.0Op+Zw=="
	scope     = "w_member_social openid profile email"
	authURL = "https://www.linkedin.com/oauth/v2/authorization"
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

func (server *Server) getLinkedinToken(ctx *gin.Context){
 	url := fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s&scope=%s&state=randomstring",
 			authURL, clientID, redirectURI, scope)
	
	ctx.Redirect(http.StatusFound,url)
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

	err = server.createSocialAccount(ctx,tokenResp)
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

func (server *Server) createSocialAccount(ctx *gin.Context, tokenResp accessTokenResponse) error {	
	// Get user id from Linkedin
	user,err := getLinkedinUserID(tokenResp)
	if err!=nil{
		return err
	}

	// Get auth_token from cookie and get username from token
	payload:= ctx.MustGet(authorizationPayload).(tk.Payload)

	// Check if the acccount already exists, if not then create a new account else update the existing account
	status,err:=server.accountExists(ctx,payload.Username)
	if !status {
		if err == sql.ErrNoRows {
			err = server.createAccount(ctx,payload.Username,user.ID,tokenResp)
			if err != nil{
				return err
			}
		}
		return err
	}else{
		err = server.updateAccount(ctx,payload.Username,user.ID,tokenResp)
		if err != nil{
			return err
		}
	}

	return nil
}

// getLinkedinUserID gets the LinkedIn user ID using the access token.
func getLinkedinUserID(tokenResp accessTokenResponse) (LinkedInUser, error) {
	var user LinkedInUser
	req, _ := http.NewRequest("GET", userIdURL, nil)
	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return  user,err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return  user,err
	}
	if err := json.Unmarshal(body, &user); err != nil {
		return  user,err
	}
	return user,nil
}

// accountExists checks if a social account already exists for a given user and platform.
func (server *Server) accountExists(ctx *gin.Context,username string) (bool, error) {
	_, err := server.connector.GetAccount(ctx, username)
	if err != nil {
		return false, err
	}
	return true, nil
}

// createAccount creates a new social account in the database.
func (server *Server) createAccount(ctx *gin.Context, username string, linkedInUserID string, tokenResp accessTokenResponse) error {
	arg := db.CreateAccountParams{
		Username:         username,
		PlatformUsername: linkedInUserID,
		AccessToken:      tokenResp.AccessToken,
		RefreshToken:     pgtype.Text{String: tokenResp.RefreshToken, Valid: true},
		ExpiresAt:        pgtype.Timestamptz{Time: time.Now().Add(time.Second * time.Duration(tokenResp.ExpiresIn)), Valid: true},
	}
	_, err := server.connector.CreateAccount(ctx, arg)
	return err
}

// updateAccount updates an existing social account in the database.
func (server *Server) updateAccount(ctx *gin.Context, username string,platformUserName string, tokenResp accessTokenResponse) error {
	updateArg := db.UpdateAccountParams{
		Username:     username,
		PlatformUsername: platformUserName,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: pgtype.Text{String: tokenResp.RefreshToken, Valid: true},
		ExpiresAt:    pgtype.Timestamptz{Time: time.Now().Add(time.Second * time.Duration(tokenResp.ExpiresIn)), Valid: true},
		UpdatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	err := server.connector.UpdateAccount(ctx, updateArg)
	return err
}