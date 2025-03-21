package server

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/loag/mcp-server-test/mcp"
)

// MCPServer represents the Model Context Protocol server
type MCPServer struct {
	Name        string
	Version     string
	Description string
	Providers   map[string]mcp.Provider
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer(name, version, description string) *MCPServer {
	return &MCPServer{
		Name:        name,
		Version:     version,
		Description: description,
		Providers:   make(map[string]mcp.Provider),
	}
}

// RegisterProvider registers a provider with the server
func (s *MCPServer) RegisterProvider(provider mcp.Provider) {
	s.Providers[provider.GetName()] = provider
}

// RegisterRoutes registers the MCP routes with the Echo instance
func (s *MCPServer) RegisterRoutes(e *echo.Echo) {
	// MCP server info endpoint
	e.GET("/", s.handleServerInfo)

	// MCP protocol endpoints
	e.POST("/v1/discover", s.handleDiscover)
	e.POST("/v1/call-tool", s.handleCallTool)
	e.POST("/v1/load-resource", s.handleLoadResource)
}

// handleServerInfo handles the server info endpoint
func (s *MCPServer) handleServerInfo(c echo.Context) error {
	info := map[string]interface{}{
		"name":        s.Name,
		"version":     s.Version,
		"description": s.Description,
		"protocol":    "mcp",
	}
	return c.JSON(http.StatusOK, info)
}

// handleDiscover handles the discover endpoint
func (s *MCPServer) handleDiscover(c echo.Context) error {
	// Create response with server capabilities
	response := mcp.DiscoverResponse{
		ServerInfo: mcp.ServerInfo{
			Name:        s.Name,
			Version:     s.Version,
			Description: s.Description,
		},
		Providers: make([]mcp.ProviderInfo, 0, len(s.Providers)),
	}

	// Add provider information
	for _, provider := range s.Providers {
		providerInfo := provider.GetInfo()
		response.Providers = append(response.Providers, providerInfo)
	}

	return c.JSON(http.StatusOK, response)
}

// handleCallTool handles the call-tool endpoint
func (s *MCPServer) handleCallTool(c echo.Context) error {
	var request mcp.CallToolRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, mcp.ErrorResponse{
			Error:   "invalid_request",
			Message: "Failed to parse request body",
		})
	}

	// Find the provider and tool
	providerName, toolName, err := parseToolID(request.ToolID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, mcp.ErrorResponse{
			Error:   "invalid_tool_id",
			Message: err.Error(),
		})
	}

	provider, exists := s.Providers[providerName]
	if !exists {
		return c.JSON(http.StatusNotFound, mcp.ErrorResponse{
			Error:   "provider_not_found",
			Message: "Provider not found: " + providerName,
		})
	}

	// Call the tool
	result, err := provider.CallTool(toolName, request)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, mcp.ErrorResponse{
			Error:   "tool_execution_error",
			Message: err.Error(),
		})
	}

	// Ensure the request ID is set
	if result.RequestID == "" {
		result.RequestID = request.RequestID
	}

	return c.JSON(http.StatusOK, result)
}

// handleLoadResource handles the load-resource endpoint
func (s *MCPServer) handleLoadResource(c echo.Context) error {
	var request mcp.LoadResourceRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, mcp.ErrorResponse{
			Error:   "invalid_request",
			Message: "Failed to parse request body",
		})
	}

	// Find the provider and resource
	providerName, resourceName, err := parseResourceID(request.ResourceID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, mcp.ErrorResponse{
			Error:   "invalid_resource_id",
			Message: err.Error(),
		})
	}

	provider, exists := s.Providers[providerName]
	if !exists {
		return c.JSON(http.StatusNotFound, mcp.ErrorResponse{
			Error:   "provider_not_found",
			Message: "Provider not found: " + providerName,
		})
	}

	// Load the resource
	result, err := provider.LoadResource(resourceName, request)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, mcp.ErrorResponse{
			Error:   "resource_load_error",
			Message: err.Error(),
		})
	}

	// Ensure the request ID is set
	if result.RequestID == "" {
		result.RequestID = request.RequestID
	}

	return c.JSON(http.StatusOK, result)
}

// Helper functions for parsing tool and resource IDs
func parseToolID(toolID string) (providerName, toolName string, err error) {
	parts := mcp.ParseID(toolID)
	if len(parts) != 2 {
		return "", "", echo.NewHTTPError(http.StatusBadRequest, "Invalid tool ID format. Expected: provider.tool")
	}
	return parts[0], parts[1], nil
}

func parseResourceID(resourceID string) (providerName, resourceName string, err error) {
	parts := mcp.ParseID(resourceID)
	if len(parts) != 2 {
		return "", "", echo.NewHTTPError(http.StatusBadRequest, "Invalid resource ID format. Expected: provider.resource")
	}
	return parts[0], parts[1], nil
}

// GenerateRequestID generates a unique request ID
func GenerateRequestID() string {
	return uuid.New().String()
}
