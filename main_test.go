package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/loag/mcp-server-test/mcp"
	"github.com/loag/mcp-server-test/server"
	"github.com/stretchr/testify/assert"
)

func setupTestServer() *echo.Echo {
	e := echo.New()
	mcpServer := server.NewMCPServer(
		"Test Filesystem MCP Server",
		"1.0.0",
		"A test MCP server implementation",
	)
	fsProvider := mcp.NewFilesystemProvider()
	mcpServer.RegisterProvider(fsProvider)
	mcpServer.RegisterRoutes(e)
	return e
}

func TestServerInfo(t *testing.T) {
	e := setupTestServer()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Test Filesystem MCP Server", response["name"])
	assert.Equal(t, "1.0.0", response["version"])
	assert.Equal(t, "A test MCP server implementation", response["description"])
	assert.Equal(t, "mcp", response["protocol"])
}

func TestDiscover(t *testing.T) {
	e := setupTestServer()
	req := httptest.NewRequest(http.MethodPost, "/v1/discover", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response mcp.DiscoverResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Test Filesystem MCP Server", response.ServerInfo.Name)
	assert.Equal(t, "1.0.0", response.ServerInfo.Version)
	assert.Equal(t, 1, len(response.Providers))
	assert.Equal(t, "filesystem", response.Providers[0].Name)
}

func TestListDirectory(t *testing.T) {
	e := setupTestServer()

	// Create a temporary test directory
	tempDir, err := os.MkdirTemp("", "mcp-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	assert.NoError(t, err)

	// Create request body
	requestBody := map[string]interface{}{
		"tool_id":    "filesystem.list",
		"request_id": "test-123",
		"params": map[string]interface{}{
			"arguments": map[string]interface{}{
				"path": tempDir,
			},
		},
	}
	jsonBody, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/call-tool", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response mcp.CallToolResult
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Print the response for debugging
	responseBytes, _ := json.MarshalIndent(response, "", "  ")
	t.Logf("Response: %s", string(responseBytes))

	assert.Equal(t, "test-123", response.RequestID)

	if response.Status == "error" && response.Error != nil {
		t.Logf("Error: %s - %s", response.Error.Code, response.Error.Message)
	}

	assert.Equal(t, "success", response.Status)

	// Verify the result contains the test file
	resultMap, ok := response.Result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "json", resultMap["type"])

	content, ok := resultMap["json"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, tempDir, content["path"])

	files, ok := content["files"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 1, len(files))

	if len(files) > 0 {
		file, ok := files[0].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "test.txt", file["name"])
	}
}

func TestReadFile(t *testing.T) {
	e := setupTestServer()

	// Create a temporary test file
	tempFile, err := os.CreateTemp("", "mcp-test-*.txt")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	// Write test content
	testContent := "Hello, MCP!"
	_, err = tempFile.WriteString(testContent)
	assert.NoError(t, err)
	tempFile.Close()

	// Create request body
	requestBody := map[string]interface{}{
		"tool_id":    "filesystem.read",
		"request_id": "test-456",
		"params": map[string]interface{}{
			"arguments": map[string]interface{}{
				"path": tempFile.Name(),
			},
		},
	}
	jsonBody, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/call-tool", bytes.NewReader(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response mcp.CallToolResult
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-456", response.RequestID)
	assert.Equal(t, "success", response.Status)

	// Verify the result contains the file content
	resultMap, ok := response.Result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "json", resultMap["type"])

	content, ok := resultMap["json"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, tempFile.Name(), content["path"])
	assert.Equal(t, testContent, content["content"])
	assert.Equal(t, true, content["is_text"])
}
