package api

import (
	"simple_bank/internal/db"

	"github.com/gin-gonic/gin"
)

// server serves HTTP requests for our banking service
type Server struct {
	store *db.Store
	router *gin.Engine
}

// NewServer creates a new HTTP server and setup routing.
func NewServer(store *db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// router
	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccounts)

	server.router = router
	return server
}

// Start runs the sHttp server on a specific address
func (server *Server) StartServer(address string) error {
	return server.router.Run(address)
}

func  errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}

