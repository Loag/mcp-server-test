package mcp

import (
	"strings"
	"time"
)

// Provider interface defines the methods that a provider must implement
type Provider interface {
	GetName() string
	GetInfo() ProviderInfo
	CallTool(toolName string, request CallToolRequest) (*CallToolResult, error)
	LoadResource(resourceName string, request LoadResourceRequest) (*LoadResourceResult, error)
}

// ServerInfo represents information about the MCP server
type ServerInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

// ProviderInfo represents information about a provider
type ProviderInfo struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Tools       []ToolInfo     `json:"tools"`
	Resources   []ResourceInfo `json:"resources"`
}

// ToolInfo represents information about a tool
type ToolInfo struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters,omitempty"`
	Returns     interface{} `json:"returns,omitempty"`
}

// ResourceInfo represents information about a resource
type ResourceInfo struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters,omitempty"`
}

// DiscoverResponse is the response from the discover endpoint
type DiscoverResponse struct {
	ServerInfo ServerInfo     `json:"server_info"`
	Providers  []ProviderInfo `json:"providers"`
}

// CallToolRequest is the request to call a tool
type CallToolRequest struct {
	ToolID    string         `json:"tool_id"`
	RequestID string         `json:"request_id"`
	Params    CallToolParams `json:"params"`
}

// CallToolParams contains the parameters for a tool call
type CallToolParams struct {
	Arguments map[string]interface{} `json:"arguments"`
}

// CallToolResult is the result of a tool call
type CallToolResult struct {
	RequestID string      `json:"request_id"`
	Status    string      `json:"status"`
	Result    interface{} `json:"result,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
}

// LoadResourceRequest is the request to load a resource
type LoadResourceRequest struct {
	ResourceID string                 `json:"resource_id"`
	RequestID  string                 `json:"request_id"`
	Params     map[string]interface{} `json:"params,omitempty"`
}

// LoadResourceResult is the result of loading a resource
type LoadResourceResult struct {
	RequestID string      `json:"request_id"`
	Status    string      `json:"status"`
	Content   interface{} `json:"content,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo represents error information
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse is a generic error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// FileInfo represents information about a file
type FileInfo struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	IsDir   bool      `json:"is_dir"`
	ModTime time.Time `json:"mod_time"`
}

// FileContent represents the content of a file
type FileContent struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	IsText  bool   `json:"is_text"`
}

// DirectoryContent represents the content of a directory
type DirectoryContent struct {
	Path  string     `json:"path"`
	Files []FileInfo `json:"files"`
}

// ParseID parses a dot-separated ID into its components
func ParseID(id string) []string {
	return strings.Split(id, ".")
}

// NewToolResultText creates a new tool result with text content
func NewToolResultText(text string) *CallToolResult {
	return &CallToolResult{
		Status: "success",
		Result: map[string]interface{}{
			"type": "text",
			"text": text,
		},
	}
}

// NewToolResultJSON creates a new tool result with JSON content
func NewToolResultJSON(data interface{}) *CallToolResult {
	return &CallToolResult{
		Status: "success",
		Result: map[string]interface{}{
			"type": "json",
			"json": data,
		},
	}
}

// NewToolResultError creates a new tool result with an error
func NewToolResultError(message string) *CallToolResult {
	return &CallToolResult{
		Status: "error",
		Error: &ErrorInfo{
			Code:    "execution_error",
			Message: message,
		},
	}
}

// NewResourceResultText creates a new resource result with text content
func NewResourceResultText(text string) *LoadResourceResult {
	return &LoadResourceResult{
		Status: "success",
		Content: map[string]interface{}{
			"type": "text",
			"text": text,
		},
	}
}

// NewResourceResultJSON creates a new resource result with JSON content
func NewResourceResultJSON(data interface{}) *LoadResourceResult {
	return &LoadResourceResult{
		Status: "success",
		Content: map[string]interface{}{
			"type": "json",
			"json": data,
		},
	}
}

// NewResourceResultError creates a new resource result with an error
func NewResourceResultError(message string) *LoadResourceResult {
	return &LoadResourceResult{
		Status: "error",
		Error: &ErrorInfo{
			Code:    "resource_error",
			Message: message,
		},
	}
}
