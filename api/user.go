package api

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/vishesh342/content-manager/db/sqlc"
	token "github.com/vishesh342/content-manager/tokens"
	util "github.com/vishesh342/content-manager/util"
)

const (
	duration    time.Duration = 30 * time.Minute
	production                = "production"
	domain                    = "localhost"
	auth_expiry               = 604800
)

type createUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func (server *Server) registerUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateUserParams{
		Username:       req.Email,
		Email:          req.Email,
		HashedPassword: hashedPassword,
		CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	user, err := server.connector.CreateUser(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, user)
}

type userInfo struct {
	Username  string             `json:"username"`
	Email     string             `json:"email"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=8"`
}
type loginUserResponse struct {
	AccessToken string   `json:"access_token"`
	User        userInfo `json:"user"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := server.connector.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	resp := util.CheckPasswordHash(req.Password, user.HashedPassword)
	if !resp {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	token, err := server.tokenMaker.CreateToken(user.Username, duration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}
	response := &loginUserResponse{
		AccessToken: token,
		User: userInfo{
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	}

	// Set an HTTP cookie to store the authentication token.
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		Domain:   domain,
		MaxAge:   auth_expiry, // 7 day
		Secure:   production == "production",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	ctx.JSON(http.StatusOK, response)
}

type getUserRequest struct {
	Username string `json:"username" uri:"username" binding:"required"`
}

func (server *Server) getUser(ctx *gin.Context) {
	var req getUserRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	authPayload := ctx.MustGet(authorizationPayload).(*token.Payload)

	if req.Username != authPayload.Username {
		err := errors.New("account does not belong to authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
	}
	user, err := server.connector.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	var userInfo userInfo
	userInfo.Username = user.Username
	userInfo.Email = user.Email
	userInfo.CreatedAt = user.CreatedAt

	ctx.JSON(http.StatusOK, userInfo)
}

type updateUserRequest struct {
	Username string `json:"username" uri:"username" binding:"required"`
	Password string `json:"password" uri:"password" binding:"required"`
}
type updateUserResponse struct {
	Remark string `json:"remark"`
}

func (server *Server) updateUser(ctx *gin.Context) {
	var req updateUserRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	authPayload := ctx.MustGet(authorizationPayload).(*token.Payload)

	if req.Username != authPayload.Username {
		err := errors.New("account does not belong to authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
	}

	arg := db.UpdateUserParams{
		Username:       req.Username,
		HashedPassword: req.Password,
		UpdatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	err := server.connector.UpdateUser(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	res := updateUserResponse{
		Remark: "user updated successfully",
	}
	ctx.JSON(http.StatusOK, res)
}
