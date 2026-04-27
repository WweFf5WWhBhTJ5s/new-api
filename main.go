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
	// Default to debug mode locally for easier development
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
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

	// Register all routes
	router.SetRouter(server)

	// Determine port
	port := os.Getenv("PORT")
	if port == "" {
		port = strconv.Itoa(common.ServerPort)
	}

	common.SysLog(fmt.Sprintf("Server listening on port %s", port))

	// Start the server
	if err := server.Run(":" + port); err != nil {
		common.FatalLog("Failed to start server: " + err.Error())
	}
}
