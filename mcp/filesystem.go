package mcp

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FilesystemProvider implements the Provider interface for filesystem operations
type FilesystemProvider struct {
	rootDir string
}

// NewFilesystemProvider creates a new filesystem provider
func NewFilesystemProvider() *FilesystemProvider {
	// Default to current directory, but this could be configurable
	return &FilesystemProvider{
		rootDir: ".",
	}
}

// GetName returns the name of the provider
func (p *FilesystemProvider) GetName() string {
	return "filesystem"
}

// GetInfo returns information about the provider
func (p *FilesystemProvider) GetInfo() ProviderInfo {
	return ProviderInfo{
		Name:        "filesystem",
		Description: "Provides access to the local filesystem",
		Tools: []ToolInfo{
			{
				ID:          "filesystem.list",
				Name:        "List Directory",
				Description: "Lists the contents of a directory",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "Path to the directory to list",
						},
					},
					"required": []string{"path"},
				},
			},
			{
				ID:          "filesystem.read",
				Name:        "Read File",
				Description: "Reads the contents of a file",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "Path to the file to read",
						},
						"encoding": map[string]interface{}{
							"type":        "string",
							"description": "Encoding of the file (text or base64)",
							"enum":        []string{"text", "base64"},
							"default":     "text",
						},
					},
					"required": []string{"path"},
				},
			},
			{
				ID:          "filesystem.write",
				Name:        "Write File",
				Description: "Writes content to a file",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "Path to the file to write",
						},
						"content": map[string]interface{}{
							"type":        "string",
							"description": "Content to write to the file",
						},
						"encoding": map[string]interface{}{
							"type":        "string",
							"description": "Encoding of the content (text or base64)",
							"enum":        []string{"text", "base64"},
							"default":     "text",
						},
					},
					"required": []string{"path", "content"},
				},
			},
			{
				ID:          "filesystem.delete",
				Name:        "Delete File",
				Description: "Deletes a file or directory",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "Path to the file or directory to delete",
						},
						"recursive": map[string]interface{}{
							"type":        "boolean",
							"description": "Whether to recursively delete directories",
							"default":     false,
						},
					},
					"required": []string{"path"},
				},
			},
		},
		Resources: []ResourceInfo{
			{
				ID:          "filesystem.file",
				Name:        "File",
				Description: "A file in the filesystem",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "Path to the file",
						},
						"encoding": map[string]interface{}{
							"type":        "string",
							"description": "Encoding of the file (text or base64)",
							"enum":        []string{"text", "base64"},
							"default":     "text",
						},
					},
					"required": []string{"path"},
				},
			},
			{
				ID:          "filesystem.directory",
				Name:        "Directory",
				Description: "A directory in the filesystem",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"path": map[string]interface{}{
							"type":        "string",
							"description": "Path to the directory",
						},
					},
					"required": []string{"path"},
				},
			},
		},
	}
}

// CallTool calls a tool provided by this provider
func (p *FilesystemProvider) CallTool(toolName string, request CallToolRequest) (*CallToolResult, error) {
	// Set the request ID in the result
	result := &CallToolResult{
		RequestID: request.RequestID,
	}

	switch toolName {
	case "list":
		return p.listDirectory(request)
	case "read":
		return p.readFile(request)
	case "write":
		return p.writeFile(request)
	case "delete":
		return p.deleteFile(request)
	default:
		result.Status = "error"
		result.Error = &ErrorInfo{
			Code:    "unknown_tool",
			Message: fmt.Sprintf("Unknown tool: %s", toolName),
		}
		return result, nil
	}
}

// LoadResource loads a resource provided by this provider
func (p *FilesystemProvider) LoadResource(resourceName string, request LoadResourceRequest) (*LoadResourceResult, error) {
	// Set the request ID in the result
	result := &LoadResourceResult{
		RequestID: request.RequestID,
	}

	switch resourceName {
	case "file":
		return p.loadFile(request)
	case "directory":
		return p.loadDirectory(request)
	default:
		result.Status = "error"
		result.Error = &ErrorInfo{
			Code:    "unknown_resource",
			Message: fmt.Sprintf("Unknown resource: %s", resourceName),
		}
		return result, nil
	}
}

// listDirectory lists the contents of a directory
func (p *FilesystemProvider) listDirectory(request CallToolRequest) (*CallToolResult, error) {
	// Get the path parameter
	pathParam, ok := request.Params.Arguments["path"].(string)
	if !ok {
		result := NewToolResultError("Path parameter is required and must be a string")
		result.RequestID = request.RequestID
		return result, nil
	}

	// Sanitize and resolve the path
	fullPath, err := p.resolvePath(pathParam)
	if err != nil {
		result := NewToolResultError(fmt.Sprintf("Invalid path: %s", err.Error()))
		result.RequestID = request.RequestID
		return result, nil
	}

	// Check if the path exists and is a directory
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			result := NewToolResultError(fmt.Sprintf("Directory not found: %s", pathParam))
			result.RequestID = request.RequestID
			return result, nil
		}
		result := NewToolResultError(fmt.Sprintf("Error accessing directory: %s", err.Error()))
		result.RequestID = request.RequestID
		return result, nil
	}

	if !info.IsDir() {
		result := NewToolResultError(fmt.Sprintf("Path is not a directory: %s", pathParam))
		result.RequestID = request.RequestID
		return result, nil
	}

	// Read the directory contents
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		result := NewToolResultError(fmt.Sprintf("Error reading directory: %s", err.Error()))
		result.RequestID = request.RequestID
		return result, nil
	}

	// Convert entries to FileInfo objects
	files := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		entryInfo, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, FileInfo{
			Name:    entry.Name(),
			Path:    filepath.Join(pathParam, entry.Name()),
			Size:    entryInfo.Size(),
			IsDir:   entry.IsDir(),
			ModTime: entryInfo.ModTime(),
		})
	}

	// Create the directory content object
	dirContent := DirectoryContent{
		Path:  pathParam,
		Files: files,
	}

	// Return the result
	result := NewToolResultJSON(dirContent)
	result.RequestID = request.RequestID
	return result, nil
}

// readFile reads the contents of a file
func (p *FilesystemProvider) readFile(request CallToolRequest) (*CallToolResult, error) {
	// Get the path parameter
	pathParam, ok := request.Params.Arguments["path"].(string)
	if !ok {
		result := NewToolResultError("Path parameter is required and must be a string")
		result.RequestID = request.RequestID
		return result, nil
	}

	// Get the encoding parameter (default to text)
	encoding := "text"
	if encodingParam, ok := request.Params.Arguments["encoding"].(string); ok {
		encoding = encodingParam
	}

	// Sanitize and resolve the path
	fullPath, err := p.resolvePath(pathParam)
	if err != nil {
		result := NewToolResultError(fmt.Sprintf("Invalid path: %s", err.Error()))
		result.RequestID = request.RequestID
		return result, nil
	}

	// Check if the path exists and is a file
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			result := NewToolResultError(fmt.Sprintf("File not found: %s", pathParam))
			result.RequestID = request.RequestID
			return result, nil
		}
		result := NewToolResultError(fmt.Sprintf("Error accessing file: %s", err.Error()))
		result.RequestID = request.RequestID
		return result, nil
	}

	if info.IsDir() {
		result := NewToolResultError(fmt.Sprintf("Path is a directory, not a file: %s", pathParam))
		result.RequestID = request.RequestID
		return result, nil
	}

	// Read the file contents
	data, err := os.ReadFile(fullPath)
	if err != nil {
		result := NewToolResultError(fmt.Sprintf("Error reading file: %s", err.Error()))
		result.RequestID = request.RequestID
		return result, nil
	}

	// Create the file content object
	var content string
	isText := true
	if encoding == "base64" {
		content = base64.StdEncoding.EncodeToString(data)
		isText = false
	} else {
		content = string(data)
	}

	fileContent := FileContent{
		Path:    pathParam,
		Content: content,
		IsText:  isText,
	}

	// Return the result
	result := NewToolResultJSON(fileContent)
	result.RequestID = request.RequestID
	return result, nil
}

// writeFile writes content to a file
func (p *FilesystemProvider) writeFile(request CallToolRequest) (*CallToolResult, error) {
	// Get the path parameter
	pathParam, ok := request.Params.Arguments["path"].(string)
	if !ok {
		result := NewToolResultError("Path parameter is required and must be a string")
		result.RequestID = request.RequestID
		return result, nil
	}

	// Get the content parameter
	contentParam, ok := request.Params.Arguments["content"].(string)
	if !ok {
		result := NewToolResultError("Content parameter is required and must be a string")
		result.RequestID = request.RequestID
		return result, nil
	}

	// Get the encoding parameter (default to text)
	encoding := "text"
	if encodingParam, ok := request.Params.Arguments["encoding"].(string); ok {
		encoding = encodingParam
	}

	// Sanitize and resolve the path
	fullPath, err := p.resolvePath(pathParam)
	if err != nil {
		result := NewToolResultError(fmt.Sprintf("Invalid path: %s", err.Error()))
		result.RequestID = request.RequestID
		return result, nil
	}

	// Create the parent directory if it doesn't exist
	parentDir := filepath.Dir(fullPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		result := NewToolResultError(fmt.Sprintf("Error creating directory: %s", err.Error()))
		result.RequestID = request.RequestID
		return result, nil
	}

	// Decode the content if necessary
	var data []byte
	if encoding == "base64" {
		data, err = base64.StdEncoding.DecodeString(contentParam)
		if err != nil {
			result := NewToolResultError(fmt.Sprintf("Error decoding base64 content: %s", err.Error()))
			result.RequestID = request.RequestID
			return result, nil
		}
	} else {
		data = []byte(contentParam)
	}

	// Write the file
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		result := NewToolResultError(fmt.Sprintf("Error writing file: %s", err.Error()))
		result.RequestID = request.RequestID
		return result, nil
	}

	// Return success
	result := NewToolResultText(fmt.Sprintf("File written successfully: %s", pathParam))
	result.RequestID = request.RequestID
	return result, nil
}

// deleteFile deletes a file or directory
func (p *FilesystemProvider) deleteFile(request CallToolRequest) (*CallToolResult, error) {
	// Get the path parameter
	pathParam, ok := request.Params.Arguments["path"].(string)
	if !ok {
		result := NewToolResultError("Path parameter is required and must be a string")
		result.RequestID = request.RequestID
		return result, nil
	}

	// Get the recursive parameter (default to false)
	recursive := false
	if recursiveParam, ok := request.Params.Arguments["recursive"].(bool); ok {
		recursive = recursiveParam
	}

	// Sanitize and resolve the path
	fullPath, err := p.resolvePath(pathParam)
	if err != nil {
		result := NewToolResultError(fmt.Sprintf("Invalid path: %s", err.Error()))
		result.RequestID = request.RequestID
		return result, nil
	}

	// Check if the path exists
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			result := NewToolResultError(fmt.Sprintf("File or directory not found: %s", pathParam))
			result.RequestID = request.RequestID
			return result, nil
		}
		result := NewToolResultError(fmt.Sprintf("Error accessing path: %s", err.Error()))
		result.RequestID = request.RequestID
		return result, nil
	}

	// Delete the file or directory
	if info.IsDir() {
		if recursive {
			if err := os.RemoveAll(fullPath); err != nil {
				result := NewToolResultError(fmt.Sprintf("Error deleting directory: %s", err.Error()))
				result.RequestID = request.RequestID
				return result, nil
			}
		} else {
			// Check if the directory is empty
			entries, err := os.ReadDir(fullPath)
			if err != nil {
				result := NewToolResultError(fmt.Sprintf("Error reading directory: %s", err.Error()))
				result.RequestID = request.RequestID
				return result, nil
			}
			if len(entries) > 0 {
				result := NewToolResultError(fmt.Sprintf("Directory is not empty: %s. Use recursive=true to delete non-empty directories", pathParam))
				result.RequestID = request.RequestID
				return result, nil
			}

			if err := os.Remove(fullPath); err != nil {
				result := NewToolResultError(fmt.Sprintf("Error deleting directory: %s", err.Error()))
				result.RequestID = request.RequestID
				return result, nil
			}
		}
	} else {
		if err := os.Remove(fullPath); err != nil {
			result := NewToolResultError(fmt.Sprintf("Error deleting file: %s", err.Error()))
			result.RequestID = request.RequestID
			return result, nil
		}
	}

	// Return success
	result := NewToolResultText(fmt.Sprintf("Successfully deleted: %s", pathParam))
	result.RequestID = request.RequestID
	return result, nil
}

// loadFile loads a file resource
func (p *FilesystemProvider) loadFile(request LoadResourceRequest) (*LoadResourceResult, error) {
	// Get the path parameter
	pathParam, ok := request.Params["path"].(string)
	if !ok {
		result := NewResourceResultError("Path parameter is required and must be a string")
		result.RequestID = request.RequestID
		return result, nil
	}

	// Get the encoding parameter (default to text)
	encoding := "text"
	if encodingParam, ok := request.Params["encoding"].(string); ok {
		encoding = encodingParam
	}

	// Sanitize and resolve the path
	fullPath, err := p.resolvePath(pathParam)
	if err != nil {
		result := NewResourceResultError(fmt.Sprintf("Invalid path: %s", err.Error()))
		result.RequestID = request.RequestID
		return result, nil
	}

	// Check if the path exists and is a file
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			result := NewResourceResultError(fmt.Sprintf("File not found: %s", pathParam))
			result.RequestID = request.RequestID
			return result, nil
		}
		result := NewResourceResultError(fmt.Sprintf("Error accessing file: %s", err.Error()))
		result.RequestID = request.RequestID
		return result, nil
	}

	if info.IsDir() {
		result := NewResourceResultError(fmt.Sprintf("Path is a directory, not a file: %s", pathParam))
		result.RequestID = request.RequestID
		return result, nil
	}

	// Read the file contents
	data, err := os.ReadFile(fullPath)
	if err != nil {
		result := NewResourceResultError(fmt.Sprintf("Error reading file: %s", err.Error()))
		result.RequestID = request.RequestID
		return result, nil
	}

	// Create the file content object
	var content string
	isText := true
	if encoding == "base64" {
		content = base64.StdEncoding.EncodeToString(data)
		isText = false
	} else {
		content = string(data)
	}

	fileContent := FileContent{
		Path:    pathParam,
		Content: content,
		IsText:  isText,
	}

	// Return the result
	result := NewResourceResultJSON(fileContent)
	result.RequestID = request.RequestID
	return result, nil
}

// loadDirectory loads a directory resource
func (p *FilesystemProvider) loadDirectory(request LoadResourceRequest) (*LoadResourceResult, error) {
	// Get the path parameter
	pathParam, ok := request.Params["path"].(string)
	if !ok {
		result := NewResourceResultError("Path parameter is required and must be a string")
		result.RequestID = request.RequestID
		return result, nil
	}

	// Sanitize and resolve the path
	fullPath, err := p.resolvePath(pathParam)
	if err != nil {
		result := NewResourceResultError(fmt.Sprintf("Invalid path: %s", err.Error()))
		result.RequestID = request.RequestID
		return result, nil
	}

	// Check if the path exists and is a directory
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			result := NewResourceResultError(fmt.Sprintf("Directory not found: %s", pathParam))
			result.RequestID = request.RequestID
			return result, nil
		}
		result := NewResourceResultError(fmt.Sprintf("Error accessing directory: %s", err.Error()))
		result.RequestID = request.RequestID
		return result, nil
	}

	if !info.IsDir() {
		result := NewResourceResultError(fmt.Sprintf("Path is not a directory: %s", pathParam))
		result.RequestID = request.RequestID
		return result, nil
	}

	// Read the directory contents
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		result := NewResourceResultError(fmt.Sprintf("Error reading directory: %s", err.Error()))
		result.RequestID = request.RequestID
		return result, nil
	}

	// Convert entries to FileInfo objects
	files := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		entryInfo, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, FileInfo{
			Name:    entry.Name(),
			Path:    filepath.Join(pathParam, entry.Name()),
			Size:    entryInfo.Size(),
			IsDir:   entry.IsDir(),
			ModTime: entryInfo.ModTime(),
		})
	}

	// Create the directory content object
	dirContent := DirectoryContent{
		Path:  pathParam,
		Files: files,
	}

	// Return the result
	result := NewResourceResultJSON(dirContent)
	result.RequestID = request.RequestID
	return result, nil
}

// resolvePath resolves and sanitizes a path
func (p *FilesystemProvider) resolvePath(path string) (string, error) {
	// If the path is absolute, use it directly
	if filepath.IsAbs(path) {
		return path, nil
	}

	// Clean the path to remove any ".." or "." components
	cleanPath := filepath.Clean(path)

	// Ensure the path doesn't try to escape the root directory
	if strings.HasPrefix(cleanPath, "..") || strings.Contains(cleanPath, "/../") {
		return "", errors.New("path attempts to access parent directory outside of root")
	}

	// Resolve the full path
	fullPath := filepath.Join(p.rootDir, cleanPath)

	// Convert to absolute path
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", err
	}

	// If we're using a relative root directory, don't enforce the prefix check
	if filepath.IsAbs(p.rootDir) {
		// Ensure the path is within the root directory
		rootAbs, err := filepath.Abs(p.rootDir)
		if err != nil {
			return "", err
		}

		if !strings.HasPrefix(absPath, rootAbs) {
			return "", errors.New("path is outside of root directory")
		}
	}

	return absPath, nil
}
