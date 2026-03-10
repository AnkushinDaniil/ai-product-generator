package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/AnkushinDaniil/memex/internal/memory"
)

// mockStorage is a mock implementation of memory.Storage for testing
type mockStorage struct{}

func (m *mockStorage) Initialize(_ context.Context) error {
	return nil
}

func (m *mockStorage) Close() error {
	return nil
}

func (m *mockStorage) Create(_ context.Context, _ *memory.Memory) error {
	return nil
}

func (m *mockStorage) Get(_ context.Context, id string) (*memory.Memory, error) {
	// Return a valid empty memory for testing
	return &memory.Memory{ID: id}, nil
}

func (m *mockStorage) Delete(_ context.Context, _ string) error {
	return nil
}

func (m *mockStorage) Search(_ context.Context, _ *memory.SearchQuery) ([]memory.SearchResult, error) {
	return nil, nil
}

func (m *mockStorage) List(_ context.Context, _ string, _ int, _ []string) ([]memory.Memory, error) {
	return nil, nil
}

func (m *mockStorage) Update(_ context.Context, _ *memory.Memory) error {
	return nil
}

func (m *mockStorage) CreateAnchor(_ context.Context, _ *memory.CodeAnchor) error {
	return nil
}

func (m *mockStorage) GetAnchorsByMemory(_ context.Context, _ string) ([]memory.CodeAnchor, error) {
	return nil, nil
}

func (m *mockStorage) FindMemoriesByAnchor(_ context.Context, _ string, _ int) ([]memory.Memory, error) {
	return nil, nil
}

func (m *mockStorage) FindMemoriesInFile(_ context.Context, _ string) ([]memory.Memory, error) {
	return nil, nil
}

func (m *mockStorage) CreateConnection(_ context.Context, _ *memory.MemoryConnection) error {
	return nil
}

func (m *mockStorage) GetConnections(_ context.Context, _ string) ([]memory.MemoryConnection, error) {
	return nil, nil
}

func (m *mockStorage) GetConnectedMemories(_ context.Context, _ string, _ int) ([]memory.Memory, error) {
	return nil, nil
}

func (m *mockStorage) MarkStale(_ context.Context, _ string, _ bool) error {
	return nil
}

func (m *mockStorage) MarkVerified(_ context.Context, _ string) error {
	return nil
}

func (m *mockStorage) GetStaleMemories(_ context.Context, _ string) ([]memory.Memory, error) {
	return nil, nil
}

// testServer creates a server with custom reader and writer for testing
func testServer(input string) (*Server, *bytes.Buffer) {
	mockStore := &mockStorage{}
	memService := memory.NewService(mockStore)
	server := NewServer(memService)

	// Replace reader with test input
	server.reader = bufio.NewReader(strings.NewReader(input))

	// Replace writer with buffer to capture output
	output := &bytes.Buffer{}
	server.writer = output

	return server, output
}

// parseResponse parses a JSON-RPC response from output
func parseResponse(t *testing.T, output *bytes.Buffer) JSONRPCResponse {
	t.Helper()

	var resp JSONRPCResponse
	if err := json.NewDecoder(output).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	return resp
}

func TestInitializeHandshake(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n"
	server, output := testServer(input)

	// Process one request
	line, err := server.reader.ReadBytes('\n')
	if err != nil {
		t.Fatalf("Failed to read input: %v", err)
	}

	var req JSONRPCRequest
	if err := json.Unmarshal(line, &req); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	server.handleRequest(&req)

	// Parse response
	resp := parseResponse(t, output)

	// Verify response structure
	if resp.JSONRPC != "2.0" {
		t.Errorf("Expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}

	if resp.ID != float64(1) {
		t.Errorf("Expected id 1, got %v", resp.ID)
	}

	if resp.Error != nil {
		t.Errorf("Expected no error, got %+v", resp.Error)
	}

	if resp.Result == nil {
		t.Fatal("Expected result, got nil")
	}

	// Verify result contains protocol version and capabilities
	resultMap, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	if resultMap["protocolVersion"] != "2024-11-05" {
		t.Errorf("Expected protocolVersion 2024-11-05, got %v", resultMap["protocolVersion"])
	}

	capabilities, ok := resultMap["capabilities"].(map[string]interface{})
	if !ok {
		t.Fatal("Capabilities is not a map")
	}

	if _, hasTools := capabilities["tools"]; !hasTools {
		t.Error("Expected capabilities to have tools")
	}

	serverInfo, ok := resultMap["serverInfo"].(map[string]interface{})
	if !ok {
		t.Fatal("ServerInfo is not a map")
	}

	if serverInfo["name"] != "memex" {
		t.Errorf("Expected server name memex, got %v", serverInfo["name"])
	}

	if serverInfo["version"] != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %v", serverInfo["version"])
	}
}

func TestToolsList(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}` + "\n"
	server, output := testServer(input)

	// Process one request
	line, err := server.reader.ReadBytes('\n')
	if err != nil {
		t.Fatalf("Failed to read input: %v", err)
	}

	var req JSONRPCRequest
	if err := json.Unmarshal(line, &req); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	server.handleRequest(&req)

	// Parse response
	resp := parseResponse(t, output)

	// Verify response
	if resp.Error != nil {
		t.Errorf("Expected no error, got %+v", resp.Error)
	}

	if resp.Result == nil {
		t.Fatal("Expected result, got nil")
	}

	// Verify result contains empty tools array (stub implementation)
	resultMap, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	tools, ok := resultMap["tools"].([]interface{})
	if !ok {
		t.Fatal("Tools is not an array")
	}

	if len(tools) != 0 {
		t.Errorf("Expected empty tools array (stub), got %d tools", len(tools))
	}
}

func TestInvalidMethod(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":3,"method":"unknown/method","params":{}}` + "\n"
	server, output := testServer(input)

	// Process one request
	line, err := server.reader.ReadBytes('\n')
	if err != nil {
		t.Fatalf("Failed to read input: %v", err)
	}

	var req JSONRPCRequest
	if err := json.Unmarshal(line, &req); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	server.handleRequest(&req)

	// Parse response
	resp := parseResponse(t, output)

	// Verify error response
	if resp.Error == nil {
		t.Fatal("Expected error, got nil")
	}

	if resp.Error.Code != MethodNotFound {
		t.Errorf("Expected error code %d (MethodNotFound), got %d", MethodNotFound, resp.Error.Code)
	}

	if resp.Error.Message != "method not found" {
		t.Errorf("Expected error message 'method not found', got %s", resp.Error.Message)
	}

	if resp.Result != nil {
		t.Errorf("Expected no result on error, got %v", resp.Result)
	}
}

func TestMalformedJSON(t *testing.T) {
	input := `{invalid json}` + "\n"
	server, output := testServer(input)

	// Process one request
	line, err := server.reader.ReadBytes('\n')
	if err != nil {
		t.Fatalf("Failed to read input: %v", err)
	}

	var req JSONRPCRequest
	if err := json.Unmarshal(line, &req); err == nil {
		t.Fatal("Expected unmarshal error for malformed JSON")
	}

	// Manually trigger error handling
	server.sendError(nil, ParseError, "parse error", "test error")

	// Parse response
	resp := parseResponse(t, output)

	// Verify error response
	if resp.Error == nil {
		t.Fatal("Expected error, got nil")
	}

	if resp.Error.Code != ParseError {
		t.Errorf("Expected error code %d (ParseError), got %d", ParseError, resp.Error.Code)
	}

	if resp.Error.Message != "parse error" {
		t.Errorf("Expected error message 'parse error', got %s", resp.Error.Message)
	}
}

func TestGracefulShutdown(t *testing.T) {
	// Create server with empty input (EOF)
	server, _ := testServer("")

	// ServeStdio should return nil on EOF
	err := server.ServeStdio()
	if err != nil {
		t.Errorf("Expected nil error on EOF, got %v", err)
	}
}

func TestToolsCallStub(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"test_tool","arguments":{}}}` + "\n"
	server, output := testServer(input)

	// Process one request
	line, err := server.reader.ReadBytes('\n')
	if err != nil {
		t.Fatalf("Failed to read input: %v", err)
	}

	var req JSONRPCRequest
	if err := json.Unmarshal(line, &req); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	server.handleRequest(&req)

	// Parse response
	resp := parseResponse(t, output)

	// Verify response (stub should return "not yet implemented")
	if resp.Error != nil {
		t.Errorf("Expected no error, got %+v", resp.Error)
	}

	if resp.Result == nil {
		t.Fatal("Expected result, got nil")
	}

	resultMap, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Result is not a map")
	}

	content, ok := resultMap["content"].([]interface{})
	if !ok {
		t.Fatal("Content is not an array")
	}

	if len(content) == 0 {
		t.Fatal("Expected content array to have elements")
	}

	firstContent, ok := content[0].(map[string]interface{})
	if !ok {
		t.Fatal("First content element is not a map")
	}

	if firstContent["type"] != "text" {
		t.Errorf("Expected content type 'text', got %v", firstContent["type"])
	}

	if firstContent["text"] != "Tool not yet implemented" {
		t.Errorf("Expected stub message, got %v", firstContent["text"])
	}
}

func TestInvalidToolCallParams(t *testing.T) {
	// Invalid params (not proper JSON object)
	input := `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":"invalid"}` + "\n"
	server, output := testServer(input)

	// Process one request
	line, err := server.reader.ReadBytes('\n')
	if err != nil {
		t.Fatalf("Failed to read input: %v", err)
	}

	var req JSONRPCRequest
	if err := json.Unmarshal(line, &req); err != nil {
		t.Fatalf("Failed to unmarshal request: %v", err)
	}

	server.handleRequest(&req)

	// Parse response
	resp := parseResponse(t, output)

	// Verify error response
	if resp.Error == nil {
		t.Fatal("Expected error for invalid params, got nil")
	}

	if resp.Error.Code != InvalidParams {
		t.Errorf("Expected error code %d (InvalidParams), got %d", InvalidParams, resp.Error.Code)
	}
}

func TestMultipleRequests(t *testing.T) {
	input := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n" +
		`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}` + "\n"

	server, output := testServer(input)

	// Process first request
	line1, err := server.reader.ReadBytes('\n')
	if err != nil {
		t.Fatalf("Failed to read first input: %v", err)
	}

	var req1 JSONRPCRequest
	if err = json.Unmarshal(line1, &req1); err != nil {
		t.Fatalf("Failed to unmarshal first request: %v", err)
	}

	server.handleRequest(&req1)

	// Process second request
	line2, err := server.reader.ReadBytes('\n')
	if err != nil {
		t.Fatalf("Failed to read second input: %v", err)
	}

	var req2 JSONRPCRequest
	if err := json.Unmarshal(line2, &req2); err != nil {
		t.Fatalf("Failed to unmarshal second request: %v", err)
	}

	server.handleRequest(&req2)

	// Parse responses using a shared decoder
	decoder := json.NewDecoder(output)

	var resp1 JSONRPCResponse
	if err := decoder.Decode(&resp1); err != nil {
		t.Fatalf("Failed to decode first response: %v", err)
	}
	if resp1.ID != float64(1) {
		t.Errorf("Expected first response ID 1, got %v", resp1.ID)
	}

	var resp2 JSONRPCResponse
	if err := decoder.Decode(&resp2); err != nil {
		t.Fatalf("Failed to decode second response: %v", err)
	}
	if resp2.ID != float64(2) {
		t.Errorf("Expected second response ID 2, got %v", resp2.ID)
	}
}

func TestReadErrorHandling(t *testing.T) {
	mockStore := &mockStorage{}
	memService := memory.NewService(mockStore)
	server := NewServer(memService)

	// Create a reader that returns an error (not EOF)
	errorReader := &errorReader{err: io.ErrUnexpectedEOF}
	server.reader = bufio.NewReader(errorReader)

	// ServeStdio should return the error
	err := server.ServeStdio()
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if !strings.Contains(err.Error(), "read error") {
		t.Errorf("Expected 'read error', got %v", err)
	}
}

// errorReader is a test helper that returns a specific error
type errorReader struct {
	err error
}

func (e *errorReader) Read(_ []byte) (n int, err error) {
	return 0, e.err
}
