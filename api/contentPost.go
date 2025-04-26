package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/vishesh342/content-manager/db/sqlc"
	"github.com/vishesh342/content-manager/util"
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

func BuildUGCPostBody(post postReqParam,)