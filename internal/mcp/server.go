package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/AnkushinDaniil/memex/internal/memory"
)

// Server represents the MCP server
type Server struct {
	memoryService *memory.Service
	reader        *bufio.Reader
	writer        io.Writer
	projectID     string // Cached project ID
}

// NewServer creates a new MCP server instance
func NewServer(memoryService *memory.Service) *Server {
	if memoryService == nil {
		panic("memoryService cannot be nil")
	}
	return &Server{
		memoryService: memoryService,
		reader:        bufio.NewReader(os.Stdin),
		writer:        os.Stdout,
		projectID:     detectProjectID(),
	}
}

// ServeStdio starts the MCP server, reading from stdin and writing to stdout
func (s *Server) ServeStdio() error {
	for {
		// Read line from stdin
		line, err := s.reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("read error: %w", err)
		}

		// Parse JSON-RPC request
		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.sendError(nil, ParseError, "parse error", err.Error())
			continue
		}

		// Handle request
		s.handleRequest(&req)
	}
}

// handleRequest processes a JSON-RPC request and sends a response
func (s *Server) handleRequest(req *JSONRPCRequest) {
	switch req.Method {
	case "initialize":
		s.handleInitialize(req)
	case "tools/list":
		s.handleListTools(req)
	case "tools/call":
		s.handleCallTool(req)
	default:
		s.sendError(req.ID, MethodNotFound, "method not found", req.Method)
	}
}

// handleInitialize handles the initialize method
func (s *Server) handleInitialize(req *JSONRPCRequest) {
	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{},
		},
		ServerInfo: ServerInfo{
			Name:    "memex",
			Version: "1.0.0",
		},
	}

	s.sendResult(req.ID, result)
}

// handleListTools handles the tools/list method
func (s *Server) handleListTools(req *JSONRPCRequest) {
	result := ListToolsResult{
		Tools: GetTools(),
	}

	s.sendResult(req.ID, result)
}

// handleCallTool handles the tools/call method
func (s *Server) handleCallTool(req *JSONRPCRequest) {
	var params CallToolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, InvalidParams, "invalid params", err.Error())
		return
	}

	// Route to appropriate tool implementation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var toolResult interface{}
	var toolErr error

	switch params.Name {
	case "memex_remember":
		toolResult, toolErr = s.executeRemember(ctx, params.Arguments)
	case "memex_recall":
		toolResult, toolErr = s.executeRecall(ctx, params.Arguments)
	case "memex_forget":
		toolResult, toolErr = s.executeForget(ctx, params.Arguments)
	case "memex_list":
		toolResult, toolErr = s.executeList(ctx, params.Arguments)
	case "memex_stats":
		toolResult, toolErr = s.executeStats(ctx, params.Arguments)
	default:
		s.sendError(req.ID, MethodNotFound, "tool not found", params.Name)
		return
	}

	// Handle tool execution result
	if toolErr != nil {
		result := CallToolResult{
			Content: []Content{
				{Type: "text", Text: fmt.Sprintf("Error: %v", toolErr)},
			},
			IsError: true,
		}
		s.sendResult(req.ID, result)
		return
	}

	// Format successful result
	resultJSON, err := json.Marshal(toolResult)
	if err != nil {
		s.sendError(req.ID, InternalError, "failed to marshal result", err.Error())
		return
	}

	result := CallToolResult{
		Content: []Content{
			{Type: "text", Text: string(resultJSON)},
		},
	}

	s.sendResult(req.ID, result)
}

// sendResult sends a successful JSON-RPC response
func (s *Server) sendResult(id, result interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal response: %v\n", err)
		return
	}
	if _, err := fmt.Fprintf(s.writer, "%s\n", data); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write response: %v\n", err)
	}
}

// sendError sends an error JSON-RPC response
func (s *Server) sendError(id interface{}, code int, message string, data interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	jsonData, err := json.Marshal(resp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal error response: %v\n", err)
		return
	}
	if _, err := fmt.Fprintf(s.writer, "%s\n", jsonData); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write error response: %v\n", err)
	}
}

// detectProjectID detects the current project identifier
func detectProjectID() string {
	// Try to get git root
	if gitRoot := getGitRoot(); gitRoot != "" {
		return filepath.Base(gitRoot)
	}

	// Fall back to current directory
	if cwd, err := os.Getwd(); err == nil {
		return filepath.Base(cwd)
	}

	return "unknown"
}

// getGitRoot returns the git repository root directory
func getGitRoot() string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil || len(output) == 0 {
		return ""
	}
	return strings.TrimSpace(string(output))
}
