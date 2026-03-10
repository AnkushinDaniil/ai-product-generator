package storage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid" //nolint:gci // import grouping

	"github.com/AnkushinDaniil/memex/internal/memory" //nolint:gci // project import
)

func setupTestDB(t *testing.T) (*SQLiteStorage, context.Context) {
	t.Helper()

	ctx := context.Background()

	// Create temp database
	dbPath := "/tmp/memex_test_" + uuid.New().String() + ".db"
	storage, err := NewSQLite(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Initialize schema
	if err := storage.Initialize(ctx); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	// Cleanup on test completion
	t.Cleanup(func() {
		_ = storage.Close()   //nolint:errcheck // test cleanup
		_ = os.Remove(dbPath) //nolint:errcheck // test cleanup
	})

	return storage, ctx
}

func TestInitialize(t *testing.T) {
	storage, ctx := setupTestDB(t)

	// Verify tables exist by attempting a query
	_, err := storage.db.ExecContext(ctx, "SELECT 1 FROM memories LIMIT 1")
	if err != nil {
		t.Errorf("memories table not created: %v", err)
	}

	_, err = storage.db.ExecContext(ctx, "SELECT 1 FROM code_anchors LIMIT 1")
	if err != nil {
		t.Errorf("code_anchors table not created: %v", err)
	}

	_, err = storage.db.ExecContext(ctx, "SELECT 1 FROM memory_connections LIMIT 1")
	if err != nil {
		t.Errorf("memory_connections table not created: %v", err)
	}
}

func TestCreateMemory(t *testing.T) {
	storage, ctx := setupTestDB(t)

	mem := &memory.Memory{
		ID:        uuid.New().String(),
		ProjectID: "test-project",
		Content:   "Test memory content",
		Type:      memory.TypeBugFix,
		Tags:      []string{"test", "bug"},
		Priority:  "high",
		IsStale:   false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := storage.Create(ctx, mem)
	if err != nil {
		t.Fatalf("Failed to create memory: %v", err)
	}

	// Retrieve and verify
	retrieved, err := storage.Get(ctx, mem.ID)
	if err != nil {
		t.Fatalf("Failed to get memory: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Memory not found")
	}

	if retrieved.Content != mem.Content {
		t.Errorf("Content mismatch: got %q, want %q", retrieved.Content, mem.Content)
	}

	if retrieved.Type != mem.Type {
		t.Errorf("Type mismatch: got %q, want %q", retrieved.Type, mem.Type)
	}
}

func TestCreateMemoryWithAnchors(t *testing.T) {
	storage, ctx := setupTestDB(t)

	// Create memory
	mem := &memory.Memory{
		ID:        uuid.New().String(),
		ProjectID: "test-project",
		Content:   "Race condition fix in session cache",
		Type:      memory.TypeBugFix,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := storage.Create(ctx, mem)
	if err != nil {
		t.Fatalf("Failed to create memory: %v", err)
	}

	// Create anchor
	anchor := &memory.CodeAnchor{
		ID:        uuid.New().String(),
		MemoryID:  mem.ID,
		File:      "internal/auth/session.go",
		Function:  "GetSession",
		StartLine: 45,
		EndLine:   67,
		GitCommit: "abc123",
	}

	err = storage.CreateAnchor(ctx, anchor)
	if err != nil {
		t.Fatalf("Failed to create anchor: %v", err)
	}

	// Retrieve anchors
	anchors, err := storage.GetAnchorsByMemory(ctx, mem.ID)
	if err != nil {
		t.Fatalf("Failed to get anchors: %v", err)
	}

	if len(anchors) != 1 {
		t.Fatalf("Expected 1 anchor, got %d", len(anchors))
	}

	if anchors[0].File != anchor.File {
		t.Errorf("File mismatch: got %q, want %q", anchors[0].File, anchor.File)
	}
}

func TestFindMemoriesByAnchor(t *testing.T) {
	storage, ctx := setupTestDB(t)

	// Create memory
	mem := &memory.Memory{
		ID:        uuid.New().String(),
		ProjectID: "test-project",
		Content:   "Bug fix at line 50",
		Type:      memory.TypeBugFix,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := storage.Create(ctx, mem)
	if err != nil {
		t.Fatalf("Failed to create memory: %v", err)
	}

	// Create anchor at lines 45-67
	anchor := &memory.CodeAnchor{
		ID:        uuid.New().String(),
		MemoryID:  mem.ID,
		File:      "internal/auth/session.go",
		StartLine: 45,
		EndLine:   67,
	}

	err = storage.CreateAnchor(ctx, anchor)
	if err != nil {
		t.Fatalf("Failed to create anchor: %v", err)
	}

	// Find memories at line 50 (within range)
	memories, err := storage.FindMemoriesByAnchor(ctx, "internal/auth/session.go", 50)
	if err != nil {
		t.Fatalf("Failed to find memories: %v", err)
	}

	if len(memories) != 1 {
		t.Fatalf("Expected 1 memory, got %d", len(memories))
	}

	if memories[0].ID != mem.ID {
		t.Errorf("Memory ID mismatch: got %q, want %q", memories[0].ID, mem.ID)
	}

	// Try line outside range
	memories, err = storage.FindMemoriesByAnchor(ctx, "internal/auth/session.go", 100)
	if err != nil {
		t.Fatalf("Failed to find memories: %v", err)
	}

	if len(memories) != 0 {
		t.Errorf("Expected 0 memories, got %d", len(memories))
	}
}

func TestFindMemoriesInFile(t *testing.T) {
	storage, ctx := setupTestDB(t)

	file := "internal/auth/session.go"

	// Create two memories with anchors in the same file
	for i := 0; i < 2; i++ {
		mem := &memory.Memory{
			ID:        uuid.New().String(),
			ProjectID: "test-project",
			Content:   "Memory " + string(rune('A'+i)),
			Type:      memory.TypeBugFix,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := storage.Create(ctx, mem)
		if err != nil {
			t.Fatalf("Failed to create memory: %v", err)
		}

		anchor := &memory.CodeAnchor{
			ID:        uuid.New().String(),
			MemoryID:  mem.ID,
			File:      file,
			StartLine: 10 + i*20,
			EndLine:   15 + i*20,
		}

		err = storage.CreateAnchor(ctx, anchor)
		if err != nil {
			t.Fatalf("Failed to create anchor: %v", err)
		}
	}

	// Find all memories in file
	memories, err := storage.FindMemoriesInFile(ctx, file)
	if err != nil {
		t.Fatalf("Failed to find memories: %v", err)
	}

	if len(memories) != 2 {
		t.Fatalf("Expected 2 memories, got %d", len(memories))
	}
}

func TestCreateConnection(t *testing.T) {
	storage, ctx := setupTestDB(t)

	// Create two memories
	mem1 := &memory.Memory{
		ID:        uuid.New().String(),
		ProjectID: "test-project",
		Content:   "Memory 1",
		Type:      memory.TypeBugFix,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mem2 := &memory.Memory{
		ID:        uuid.New().String(),
		ProjectID: "test-project",
		Content:   "Memory 2",
		Type:      memory.TypeConnection,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := storage.Create(ctx, mem1); err != nil {
		t.Fatalf("Failed to create memory 1: %v", err)
	}
	if err := storage.Create(ctx, mem2); err != nil {
		t.Fatalf("Failed to create memory 2: %v", err)
	}

	// Create connection
	conn := &memory.MemoryConnection{
		ID:           uuid.New().String(),
		FromMemoryID: mem1.ID,
		ToMemoryID:   mem2.ID,
		Relationship: memory.ConnAffects,
		Description:  "Memory 1 affects Memory 2",
		CreatedAt:    time.Now(),
	}

	err := storage.CreateConnection(ctx, conn)
	if err != nil {
		t.Fatalf("Failed to create connection: %v", err)
	}

	// Get connections
	connections, err := storage.GetConnections(ctx, mem1.ID)
	if err != nil {
		t.Fatalf("Failed to get connections: %v", err)
	}

	if len(connections) != 1 {
		t.Fatalf("Expected 1 connection, got %d", len(connections))
	}

	if connections[0].Relationship != memory.ConnAffects {
		t.Errorf("Relationship mismatch: got %q, want %q", connections[0].Relationship, memory.ConnAffects)
	}
}

func TestGetConnectedMemories(t *testing.T) {
	storage, ctx := setupTestDB(t)

	// Create chain: mem1 -> mem2 -> mem3
	mem1 := createTestMemory(ctx, storage, "Memory 1")
	mem2 := createTestMemory(ctx, storage, "Memory 2")
	mem3 := createTestMemory(ctx, storage, "Memory 3")

	// Create connections
	conn1 := &memory.MemoryConnection{
		ID:           uuid.New().String(),
		FromMemoryID: mem1.ID,
		ToMemoryID:   mem2.ID,
		Relationship: memory.ConnAffects,
		CreatedAt:    time.Now(),
	}

	conn2 := &memory.MemoryConnection{
		ID:           uuid.New().String(),
		FromMemoryID: mem2.ID,
		ToMemoryID:   mem3.ID,
		Relationship: memory.ConnDependsOn,
		CreatedAt:    time.Now(),
	}

	if err := storage.CreateConnection(ctx, conn1); err != nil {
		t.Fatalf("Failed to create connection 1: %v", err)
	}
	if err := storage.CreateConnection(ctx, conn2); err != nil {
		t.Fatalf("Failed to create connection 2: %v", err)
	}

	// Get connected memories with depth 1
	connected, err := storage.GetConnectedMemories(ctx, mem1.ID, 1)
	if err != nil {
		t.Fatalf("Failed to get connected memories: %v", err)
	}

	if len(connected) != 1 {
		t.Fatalf("Expected 1 connected memory at depth 1, got %d", len(connected))
	}

	// Get connected memories with depth 2
	connected, err = storage.GetConnectedMemories(ctx, mem1.ID, 2)
	if err != nil {
		t.Fatalf("Failed to get connected memories: %v", err)
	}

	if len(connected) != 2 {
		t.Fatalf("Expected 2 connected memories at depth 2, got %d", len(connected))
	}
}

func TestMarkStale(t *testing.T) {
	storage, ctx := setupTestDB(t)

	mem := createTestMemory(ctx, storage, "Test memory")

	// Mark as stale
	err := storage.MarkStale(ctx, mem.ID, true)
	if err != nil {
		t.Fatalf("Failed to mark stale: %v", err)
	}

	// Retrieve and verify
	retrieved, err := storage.Get(ctx, mem.ID)
	if err != nil {
		t.Fatalf("Failed to get memory: %v", err)
	}

	if !retrieved.IsStale {
		t.Error("Memory should be marked as stale")
	}

	// Mark as not stale
	err = storage.MarkStale(ctx, mem.ID, false)
	if err != nil {
		t.Fatalf("Failed to mark not stale: %v", err)
	}

	retrieved, err = storage.Get(ctx, mem.ID)
	if err != nil {
		t.Fatalf("Failed to get memory: %v", err)
	}

	if retrieved.IsStale {
		t.Error("Memory should not be marked as stale")
	}
}

func TestMarkVerified(t *testing.T) {
	storage, ctx := setupTestDB(t)

	mem := createTestMemory(ctx, storage, "Test memory")

	// Mark as stale first
	if err := storage.MarkStale(ctx, mem.ID, true); err != nil {
		t.Fatalf("Failed to mark stale: %v", err)
	}

	// Mark as verified
	err := storage.MarkVerified(ctx, mem.ID)
	if err != nil {
		t.Fatalf("Failed to mark verified: %v", err)
	}

	// Retrieve and verify
	retrieved, err := storage.Get(ctx, mem.ID)
	if err != nil {
		t.Fatalf("Failed to get memory: %v", err)
	}

	if retrieved.IsStale {
		t.Error("Memory should not be stale after verification")
	}

	if retrieved.LastVerified == nil {
		t.Error("LastVerified should be set")
	}
}

func TestGetStaleMemories(t *testing.T) {
	storage, ctx := setupTestDB(t)

	projectID := "test-project"

	// Create mix of stale and fresh memories
	mem1 := createTestMemory(ctx, storage, "Stale memory 1")
	mem2 := createTestMemory(ctx, storage, "Fresh memory")
	mem3 := createTestMemory(ctx, storage, "Stale memory 2")

	if err := storage.MarkStale(ctx, mem1.ID, true); err != nil {
		t.Fatalf("Failed to mark mem1 stale: %v", err)
	}
	if err := storage.MarkStale(ctx, mem3.ID, true); err != nil {
		t.Fatalf("Failed to mark mem3 stale: %v", err)
	}

	// Get stale memories
	stale, err := storage.GetStaleMemories(ctx, projectID)
	if err != nil {
		t.Fatalf("Failed to get stale memories: %v", err)
	}

	if len(stale) != 2 {
		t.Fatalf("Expected 2 stale memories, got %d", len(stale))
	}

	// Verify fresh memory is not included
	for _, m := range stale {
		if m.ID == mem2.ID {
			t.Error("Fresh memory should not be in stale list")
		}
	}
}

func TestUpdate(t *testing.T) {
	storage, ctx := setupTestDB(t)

	mem := createTestMemory(ctx, storage, "Original content")

	// Update content
	mem.Content = "Updated content"
	mem.UpdatedAt = time.Now()

	err := storage.Update(ctx, mem)
	if err != nil {
		t.Fatalf("Failed to update memory: %v", err)
	}

	// Retrieve and verify
	retrieved, err := storage.Get(ctx, mem.ID)
	if err != nil {
		t.Fatalf("Failed to get memory: %v", err)
	}

	if retrieved.Content != "Updated content" {
		t.Errorf("Content not updated: got %q, want %q", retrieved.Content, "Updated content")
	}
}

func TestDelete(t *testing.T) {
	storage, ctx := setupTestDB(t)

	mem := createTestMemory(ctx, storage, "To be deleted")

	// Delete memory
	err := storage.Delete(ctx, mem.ID)
	if err != nil {
		t.Fatalf("Failed to delete memory: %v", err)
	}

	// Verify it's gone
	retrieved, err := storage.Get(ctx, mem.ID)
	if err == nil || err.Error() != "memory not found" {
		t.Fatalf("Expected 'memory not found' error, got: %v", err)
	}

	if retrieved != nil {
		t.Error("Memory should be deleted")
	}
}

func TestSearch(t *testing.T) {
	storage, ctx := setupTestDB(t)

	// Create memories with searchable content
	createTestMemory(ctx, storage, "Authentication bug fix with JWT tokens")
	createTestMemory(ctx, storage, "Cache optimization for session storage")
	createTestMemory(ctx, storage, "JWT token validation improvements")

	// Search for "JWT"
	query := &memory.SearchQuery{
		Query:     "JWT",
		ProjectID: "test-project",
		Limit:     10,
	}

	results, err := storage.Search(ctx, query)
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results for 'JWT', got %d", len(results))
	}
}

func TestList(t *testing.T) {
	storage, ctx := setupTestDB(t)

	// Create several memories
	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
		createTestMemory(ctx, storage, "Memory "+string(rune('A'+i)))
	}

	// List with limit
	memories, err := storage.List(ctx, "test-project", 3, nil)
	if err != nil {
		t.Fatalf("Failed to list memories: %v", err)
	}

	if len(memories) != 3 {
		t.Fatalf("Expected 3 memories, got %d", len(memories))
	}

	// Verify they're ordered by created_at DESC (most recent first)
	for i := 0; i < len(memories)-1; i++ {
		if memories[i].CreatedAt.Before(memories[i+1].CreatedAt) {
			t.Error("Memories should be ordered by created_at DESC")
		}
	}
}

// Helper functions

func createTestMemory(ctx context.Context, storage *SQLiteStorage, content string) *memory.Memory {
	mem := &memory.Memory{
		ID:        uuid.New().String(),
		ProjectID: "test-project",
		Content:   content,
		Type:      memory.TypeGeneral,
		Priority:  "normal",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_ = storage.Create(ctx, mem) //nolint:errcheck // test helper
	return mem
}
