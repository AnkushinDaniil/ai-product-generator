package mcp

import (
	"context"
	"testing"

	"github.com/AnkushinDaniil/memex/internal/memory"
	"github.com/AnkushinDaniil/memex/internal/storage"
)

// setupTestServer creates a test server with in-memory SQLite
func setupTestServer(t *testing.T) *Server {
	t.Helper()

	store, err := storage.NewSQLite(":memory:")
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	err = store.Initialize(context.Background())
	if err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	memService := memory.NewService(store)
	server := NewServer(memService)

	t.Cleanup(func() {
		_ = store.Close()
	})

	return server
}

func TestRememberBasic(t *testing.T) {
	server := setupTestServer(t)

	params := map[string]interface{}{
		"content": "Test memory content",
	}

	result, err := server.executeRemember(context.Background(), params)
	if err != nil {
		t.Fatalf("executeRemember failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if resultMap["memory_id"] == "" {
		t.Error("Expected memory_id to be set")
	}
}

func TestRememberWithTags(t *testing.T) {
	server := setupTestServer(t)

	params := map[string]interface{}{
		"content": "Test memory with tags",
		"tags":    []interface{}{"tag1", "tag2"},
	}

	result, err := server.executeRemember(context.Background(), params)
	if err != nil {
		t.Fatalf("executeRemember failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if resultMap["memory_id"] == "" {
		t.Error("Expected memory_id to be set")
	}

	// Verify memory was created with tags
	listParams := map[string]interface{}{}
	listResult, err := server.executeList(context.Background(), listParams)
	if err != nil {
		t.Fatalf("executeList failed: %v", err)
	}

	listMap := listResult.(map[string]interface{})
	memories := listMap["memories"].([]map[string]interface{})

	if len(memories) != 1 {
		t.Errorf("Expected 1 memory, got %d", len(memories))
	}
}

func TestRememberMissingContent(t *testing.T) {
	server := setupTestServer(t)

	params := map[string]interface{}{}

	_, err := server.executeRemember(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing content, got nil")
	}
}

func TestRecallFound(t *testing.T) {
	server := setupTestServer(t)

	// First create a memory
	rememberParams := map[string]interface{}{
		"content": "Test memory",
	}
	_, err := server.executeRemember(context.Background(), rememberParams)
	if err != nil {
		t.Fatalf("executeRemember failed: %v", err)
	}

	params := map[string]interface{}{
		"query": "test",
	}

	result, err := server.executeRecall(context.Background(), params)
	if err != nil {
		t.Fatalf("executeRecall failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	memories := resultMap["memories"].([]map[string]interface{})

	if len(memories) != 1 {
		t.Errorf("Expected 1 memory, got %d", len(memories))
	}
}

func TestRecallNotFound(t *testing.T) {
	server := setupTestServer(t)

	params := map[string]interface{}{
		"query": "nonexistent",
	}

	result, err := server.executeRecall(context.Background(), params)
	if err != nil {
		t.Fatalf("executeRecall failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	memories := resultMap["memories"].([]map[string]interface{})

	if len(memories) != 0 {
		t.Errorf("Expected 0 memories, got %d", len(memories))
	}
}

func TestRecallWithLimit(t *testing.T) {
	server := setupTestServer(t)

	params := map[string]interface{}{
		"query": "test",
		"limit": float64(5),
	}

	_, err := server.executeRecall(context.Background(), params)
	if err != nil {
		t.Fatalf("executeRecall failed: %v", err)
	}
}

func TestForgetExisting(t *testing.T) {
	server := setupTestServer(t)

	params := map[string]interface{}{
		"memory_id": "test-id",
	}

	result, err := server.executeForget(context.Background(), params)
	if err != nil {
		t.Fatalf("executeForget failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	if !resultMap["success"].(bool) {
		t.Error("Expected success to be true")
	}
}

func TestForgetMissing(t *testing.T) {
	server := setupTestServer(t)

	params := map[string]interface{}{}

	_, err := server.executeForget(context.Background(), params)
	if err == nil {
		t.Error("Expected error for missing memory_id, got nil")
	}
}

func TestListEmpty(t *testing.T) {
	server := setupTestServer(t)

	params := map[string]interface{}{}

	result, err := server.executeList(context.Background(), params)
	if err != nil {
		t.Fatalf("executeList failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	memories := resultMap["memories"].([]map[string]interface{})

	if len(memories) != 0 {
		t.Errorf("Expected 0 memories, got %d", len(memories))
	}
}

func TestListWithMemories(t *testing.T) {
	server := setupTestServer(t)

	// Create test memories
	_, _ = server.executeRemember(context.Background(), map[string]interface{}{"content": "Memory 1"})
	_, _ = server.executeRemember(context.Background(), map[string]interface{}{"content": "Memory 2"})

	params := map[string]interface{}{}

	result, err := server.executeList(context.Background(), params)
	if err != nil {
		t.Fatalf("executeList failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	memories := resultMap["memories"].([]map[string]interface{})

	if len(memories) != 2 {
		t.Errorf("Expected 2 memories, got %d", len(memories))
	}
}

func TestStats(t *testing.T) {
	server := setupTestServer(t)

	// Create test memories
	_, _ = server.executeRemember(context.Background(), map[string]interface{}{"content": "Memory 1", "type": "general"})
	_, _ = server.executeRemember(context.Background(), map[string]interface{}{"content": "Memory 2", "type": "bug-fix"})

	params := map[string]interface{}{}

	result, err := server.executeStats(context.Background(), params)
	if err != nil {
		t.Fatalf("executeStats failed: %v", err)
	}

	resultMap := result.(map[string]interface{})
	totalMemories := resultMap["total_memories"].(int)

	if totalMemories != 2 {
		t.Errorf("Expected 2 total memories, got %d", totalMemories)
	}
}
