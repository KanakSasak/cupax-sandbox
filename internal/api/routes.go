package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(handler *Handler, frontendDir string) *gin.Engine {
	router := gin.Default()

	// CORS configuration for development
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:5173", "http://localhost:3000", "http://localhost:8080"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	router.Use(cors.New(config))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Analysis endpoints
		v1.POST("/analyze", handler.HandleUploadFile)
		v1.GET("/analyses", handler.HandleGetAnalyses)
		v1.GET("/analyses/:id", handler.HandleGetAnalysisByID)

		// Whitelist endpoints
		v1.POST("/whitelists", handler.HandleCreateWhitelist)
		v1.GET("/whitelists", handler.HandleGetWhitelists)
		v1.GET("/whitelists/:id", handler.HandleGetWhitelistByID)
		v1.PUT("/whitelists/:id", handler.HandleUpdateWhitelist)
		v1.DELETE("/whitelists/:id", handler.HandleDeleteWhitelist)
		v1.POST("/whitelists/bulk", handler.HandleBulkCreateWhitelists)
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Serve frontend static files (if frontend directory is provided)
	if frontendDir != "" {
		router.Static("/assets", frontendDir+"/assets")
		router.StaticFile("/", frontendDir+"/index.html")
		router.NoRoute(func(c *gin.Context) {
			c.File(frontendDir + "/index.html")
		})
	}

	return router
}
