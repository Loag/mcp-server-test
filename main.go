package main

import (
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/loag/mcp-server-test/mcp"
	"github.com/loag/mcp-server-test/server"
)

func main() {
	// Create a new Echo instance
	e := echo.New()

	// Add middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Create MCP server
	mcpServer := server.NewMCPServer(
		"Filesystem MCP Server",
		"1.0.0",
		"A Model Context Protocol server implementation that provides access to the local file system",
	)

	// Register filesystem tools
	fsProvider := mcp.NewFilesystemProvider()
	mcpServer.RegisterProvider(fsProvider)

	// Setup MCP routes
	mcpServer.RegisterRoutes(e)

	// Determine port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Starting MCP server on port %s", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
