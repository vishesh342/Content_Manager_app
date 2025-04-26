package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/vishesh342/content-manager/db/sqlc"
	tk "github.com/vishesh342/content-manager/tokens"
	"github.com/vishesh342/content-manager/util"
)
const (
	postUrl   =  "https://api.linkedin.com/v2/ugcPosts"
)

type postReqParam struct {
	Content       string    `json:"content" binding:"required"`
	MediaType     string    `json:"media_type" binding:"required"`
	MediaUrns     []byte    `json:"media_urns"`
	ScheduledTime time.Time `json:"scheduled_time" binding:"required"`
	Visibility    string    `json:"visibility" binding:"required"`
	AccountID     string     `json:"account_id" binding:"required"`
}

func (server *Server) contentPostHandler(ctx *gin.Context) {
	var req postReqParam
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest,errorResponse(err))
		return
	}
	if req.ScheduledTime.IsZero() {

	}

	id:=util.GeneratePostID(req.AccountID,req.ScheduledTime)
	post:=db.CreatePostParams{
		ID:			id,
		Content:		req.Content,
		MediaType:		req.MediaType,
		MediaUrns:		req.MediaUrns,
		ScheduledTime:	pgtype.Timestamptz{Time: req.ScheduledTime, Valid: true},
		Visibility:	req.Visibility,
		AccountID:	req.AccountID,
	}

	_, err := server.connector.CreatePost(ctx, post)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

}

func BuildUGCPostBodyasd (post postReqParam,authID string) map[string]any {
	mediaList := []map[string]any{}
	if post.MediaType != "NONE" && post.MediaType != "ARTICLE" && len(post.MediaUrns) > 0 { 
		for _, urn := range post.MediaUrns {
			mediaList = append(mediaList, map[string]interface{}{
				"status": "READY",
				"media":  urn,
			})
		}
	}	
	authorURN := "urn:li:person:" + authID
	ugcBody := map[string]interface{}{
		"author":         authorURN, 
		"lifecycleState": "PUBLISHED",
		"specificContent": map[string]interface{}{
			"com.linkedin.ugc.ShareContent": map[string]any{
				"shareCommentary": map[string]string{
					"text": post.Content,
				},
				"shareMediaCategory": post.MediaType,
			},
		},
		"visibility": map[string]any{
			"com.linkedin.ugc.MemberNetworkVisibility": post.Visibility,
		},
	}
	if len(mediaList) > 0 {
		ugcContent := ugcBody["specificContent"].(map[string]any)["com.linkedin.ugc.ShareContent"].(map[string]any)
		ugcContent["media"] = mediaList
	}
	return ugcBody
}	


func(server *Server) Postcontent(ctx *gin.Context,accountID int, ugcBody map[string]interface{})error{
	payload:= ctx.MustGet(authorizationPayload).(tk.Payload)
	
	account,err := server.connector.GetAccount(ctx, payload.Username) 
	if err != nil {
		return err
	}

	bodyBytes, err := json.Marshal(ugcBody)
	if err != nil {
		return fmt.Errorf("failed to marshal post body: %w", err)
	}

	req, err := http.NewRequest("POST",postUrl, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create LinkedIn API request: %w", err)
	}
	accessToken := "Bearer "+account.AccessToken

	req.Header.Set("Authorization", accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Restli-Protocol-Version", "2.0.0")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("linkedin API call failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-success status codes [49]
	var responseBody map[string]interface{}
	// Attempt to decode the error response body for more details
	if decodeErr := json.NewDecoder(resp.Body).Decode(&responseBody); decodeErr != nil {
		return fmt.Errorf("linkedin API returned status %d (failed to decode error response)", resp.StatusCode)
	}

	return nil
}