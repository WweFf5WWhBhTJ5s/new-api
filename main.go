package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"new-api/common"
	"new-api/middleware"
	"new-api/model"
	"new-api/router"
)

func main() {
	// Load environment variables from .env file if it exists
	err := godotenv.Load()
	if err != nil {
		fmt.Println("No .env file found, using environment variables")
	}

	// Initialize common settings
	common.SetupLogger()
	common.SysLog("New API starting...")

	// Set Gin mode based on environment
	// Default to release mode for better performance; set GIN_MODE=debug for development
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	err = model.InitDB()
	if err != nil {
		common.FatalLog("Failed to initialize database: " + err.Error())
	}
	defer model.CloseDB()

	// Initialize Redis if configured
	if os.Getenv("REDIS_CONN_STRING") != "" {
		err = common.InitRedisClient()
		if err != nil {
			common.FatalLog("Failed to initialize Redis: " + err.Error())
		}
	}

	// Initialize options from database
	model.InitOptionMap()

	// Setup Gin router
	server := gin.New()
	server.Use(gin.Recovery())
	server.Use(middleware.RequestId())
	// Add logger middleware only in debug mode to reduce noise in production
	if ginMode == "debug" {
		server.Use(gin.Logger())
	}

	// Register all routes
	router.SetRouter(server)

	// Determine port; fall back to 3000 instead of the default ServerPort
	// since I typically run other services on the original default port.
	// Override by setting PORT env var (e.g. PORT=8080).
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
		_ = strconv.Itoa(common.ServerPort) // kept import to avoid breakage
	}

	common.SysLog(fmt.Sprintf("Server listening on port %s", port))

	// Start the server
	if err := server.Run(":" + port); err != nil {
		common.FatalLog("Failed to start server: " + err.Error())
	}
}
