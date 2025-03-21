# Model Context Protocol (MCP) Server - Filesystem Implementation

This is an educational implementation of a Model Context Protocol (MCP) server in Go using the Echo framework. The server provides access to the local filesystem through the MCP protocol.

## Overview

The Model Context Protocol (MCP) is an open standard that enables developers to build secure, two-way connections between their data sources and AI-powered tools. This implementation demonstrates how to create an MCP server that provides access to the local filesystem.

## Features

- **Filesystem Provider**: Provides access to the local filesystem through MCP tools and resources
- **Tools**:
  - `filesystem.list`: Lists the contents of a directory
  - `filesystem.read`: Reads the contents of a file
  - `filesystem.write`: Writes content to a file
  - `filesystem.delete`: Deletes a file or directory
- **Resources**:
  - `filesystem.file`: Represents a file in the filesystem
  - `filesystem.directory`: Represents a directory in the filesystem

## Security Considerations

This implementation includes basic path sanitization to prevent directory traversal attacks, but it is intended for educational purposes only. In a production environment, additional security measures would be necessary, such as:

- User authentication and authorization
- More robust input validation
- Rate limiting
- Audit logging
- Sandboxing

## Getting Started

### Prerequisites

- Go 1.24 or higher

### Installation

1. Clone the repository
2. Install dependencies: `go mod tidy`
3. Run the server: `go run main.go`

The server will start on port 8080 by default. You can change the port by setting the `PORT` environment variable.

## API Endpoints

- `GET /`: Server information
- `POST /v1/discover`: Discover server capabilities
- `POST /v1/call-tool`: Call a tool
- `POST /v1/load-resource`: Load a resource

## Example Usage

### Discover Server Capabilities

```bash
curl -X POST http://localhost:8080/v1/discover
```

### List Directory Contents

```bash
curl -X POST http://localhost:8080/v1/call-tool \
  -H "Content-Type: application/json" \
  -d '{
    "tool_id": "filesystem.list",
    "request_id": "req-123",
    "params": {
      "arguments": {
        "path": "."
      }
    }
  }'
```

### Read File Contents

```bash
curl -X POST http://localhost:8080/v1/call-tool \
  -H "Content-Type: application/json" \
  -d '{
    "tool_id": "filesystem.read",
    "request_id": "req-123",
    "params": {
      "arguments": {
        "path": "README.md"
      }
    }
  }'
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Disclaimer

This implementation is for educational purposes only and should not be used in production environments without proper security review and hardening.