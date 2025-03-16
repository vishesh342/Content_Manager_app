package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/vishesh342/content-manager/db/sqlc"
	tk "github.com/vishesh342/content-manager/tokens"
)

const(
	symmetricKey = "12345678901234567890123456789012"
)

type Server struct{
	connector *db.DBConnector
	tokenMaker tk.TokenMaker
	router *gin.Engine
}

func NewServer(dbConn *pgxpool.Pool) (*Server,error) {
	maker,err:= tk.NewToken(symmetricKey)
	if err!=nil{
		return nil,fmt.Errorf("failed to create token maker: %w",err)
	}
	server := &Server{}
	server.tokenMaker = maker
	server.connector = db.NewConnector(dbConn)
	router := gin.Default()

	authGroup:=router.Group("/").Use(authMiddleware(server.tokenMaker))
	router.POST("/account", server.registerUser)
	router.POST("/account/login", server.loginUser)
	router.GET("/oauth/linkedin/callback",server.handleLinkedInCallback)
	
	
	// added to Authorization Group.
	authGroup.GET("oauth/linkedin",server.getLinkedinToken)
	authGroup.GET("/account/:username", server.getUser)
	authGroup.PUT("/account", server.updateUser)

	server.router = router
	return server,nil
}

func (server *Server) Run(add string)error{
	return server.router.Run(add)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}