package api

import (
	"simple_bank/internal/db"

	"github.com/gin-gonic/gin"
)

// Server is an HTTP server that handles banking API requests.
// It uses a Gin router for HTTP routing and a database store for data persistence.
type Server struct {
	store  db.Store
	router *gin.Engine
}

// NewServer creates a new HTTP server and setup routing.
// NewServer creates and returns a new Server instance with the provided database store.
// It initializes a Gin router and sets up the following routes:
// - POST /accounts: creates a new account
// - GET /accounts/:id: retrieves an account by ID
// - GET /accounts: lists all accounts
func NewServer(store db.Store) *Server {
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

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
