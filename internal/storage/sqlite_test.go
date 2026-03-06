package storage

import (
	"os"
	"testing"
	"time"

	"github.com/AnkushinDaniil/memex/internal/memory"
	"github.com/google/uuid"
)

func setupTestDB(t *testing.T) *SQLiteStorage {
	t.Helper()

	// Create temp database
	dbPath := "/tmp/memex_test_" + uuid.New().String() + ".db"
	storage, err := NewSQLite(dbPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Initialize schema
	if err := storage.Initialize(); err != nil {
		t.Fatalf("Failed to initialize storage: %v", err)
	}

	// Cleanup on test completion
	t.Cleanup(func() {
		storage.Close()
		os.Remove(dbPath)
	})

	return storage
}

func TestInitialize(t *testing.T) {
	storage := setupTestDB(t)

	// Verify tables exist by attempting a query
	_, err := storage.db.Exec("SELECT 1 FROM memories LIMIT 1")
	if err != nil {
		t.Errorf("memories table not created: %v", err)
	}

	_, err = storage.db.Exec("SELECT 1 FROM code_anchors LIMIT 1")
	if err != nil {
		t.Errorf("code_anchors table not created: %v", err)
	}

	_, err = storage.db.Exec("SELECT 1 FROM memory_connections LIMIT 1")
	if err != nil {
		t.Errorf("memory_connections table not created: %v", err)
	}
}

func TestCreateMemory(t *testing.T) {
	storage := setupTestDB(t)

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

	err := storage.Create(mem)
	if err != nil {
		t.Fatalf("Failed to create memory: %v", err)
	}

	// Retrieve and verify
	retrieved, err := storage.Get(mem.ID)
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
	storage := setupTestDB(t)

	// Create memory
	mem := &memory.Memory{
		ID:        uuid.New().String(),
		ProjectID: "test-project",
		Content:   "Race condition fix in session cache",
		Type:      memory.TypeBugFix,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := storage.Create(mem)
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

	err = storage.CreateAnchor(anchor)
	if err != nil {
		t.Fatalf("Failed to create anchor: %v", err)
	}

	// Retrieve anchors
	anchors, err := storage.GetAnchorsByMemory(mem.ID)
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
	storage := setupTestDB(t)

	// Create memory
	mem := &memory.Memory{
		ID:        uuid.New().String(),
		ProjectID: "test-project",
		Content:   "Bug fix at line 50",
		Type:      memory.TypeBugFix,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := storage.Create(mem)
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

	err = storage.CreateAnchor(anchor)
	if err != nil {
		t.Fatalf("Failed to create anchor: %v", err)
	}

	// Find memories at line 50 (within range)
	memories, err := storage.FindMemoriesByAnchor("internal/auth/session.go", 50)
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
	memories, err = storage.FindMemoriesByAnchor("internal/auth/session.go", 100)
	if err != nil {
		t.Fatalf("Failed to find memories: %v", err)
	}

	if len(memories) != 0 {
		t.Errorf("Expected 0 memories, got %d", len(memories))
	}
}

func TestFindMemoriesInFile(t *testing.T) {
	storage := setupTestDB(t)

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

		err := storage.Create(mem)
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

		err = storage.CreateAnchor(anchor)
		if err != nil {
			t.Fatalf("Failed to create anchor: %v", err)
		}
	}

	// Find all memories in file
	memories, err := storage.FindMemoriesInFile(file)
	if err != nil {
		t.Fatalf("Failed to find memories: %v", err)
	}

	if len(memories) != 2 {
		t.Fatalf("Expected 2 memories, got %d", len(memories))
	}
}

func TestCreateConnection(t *testing.T) {
	storage := setupTestDB(t)

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

	storage.Create(mem1)
	storage.Create(mem2)

	// Create connection
	conn := &memory.MemoryConnection{
		ID:           uuid.New().String(),
		FromMemoryID: mem1.ID,
		ToMemoryID:   mem2.ID,
		Relationship: memory.ConnAffects,
		Description:  "Memory 1 affects Memory 2",
		CreatedAt:    time.Now(),
	}

	err := storage.CreateConnection(conn)
	if err != nil {
		t.Fatalf("Failed to create connection: %v", err)
	}

	// Get connections
	connections, err := storage.GetConnections(mem1.ID)
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
	storage := setupTestDB(t)

	// Create chain: mem1 -> mem2 -> mem3
	mem1 := createTestMemory(storage, "Memory 1")
	mem2 := createTestMemory(storage, "Memory 2")
	mem3 := createTestMemory(storage, "Memory 3")

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

	storage.CreateConnection(conn1)
	storage.CreateConnection(conn2)

	// Get connected memories with depth 1
	connected, err := storage.GetConnectedMemories(mem1.ID, 1)
	if err != nil {
		t.Fatalf("Failed to get connected memories: %v", err)
	}

	if len(connected) != 1 {
		t.Fatalf("Expected 1 connected memory at depth 1, got %d", len(connected))
	}

	// Get connected memories with depth 2
	connected, err = storage.GetConnectedMemories(mem1.ID, 2)
	if err != nil {
		t.Fatalf("Failed to get connected memories: %v", err)
	}

	if len(connected) != 2 {
		t.Fatalf("Expected 2 connected memories at depth 2, got %d", len(connected))
	}
}

func TestMarkStale(t *testing.T) {
	storage := setupTestDB(t)

	mem := createTestMemory(storage, "Test memory")

	// Mark as stale
	err := storage.MarkStale(mem.ID, true)
	if err != nil {
		t.Fatalf("Failed to mark stale: %v", err)
	}

	// Retrieve and verify
	retrieved, err := storage.Get(mem.ID)
	if err != nil {
		t.Fatalf("Failed to get memory: %v", err)
	}

	if !retrieved.IsStale {
		t.Error("Memory should be marked as stale")
	}

	// Mark as not stale
	err = storage.MarkStale(mem.ID, false)
	if err != nil {
		t.Fatalf("Failed to mark not stale: %v", err)
	}

	retrieved, err = storage.Get(mem.ID)
	if err != nil {
		t.Fatalf("Failed to get memory: %v", err)
	}

	if retrieved.IsStale {
		t.Error("Memory should not be marked as stale")
	}
}

func TestMarkVerified(t *testing.T) {
	storage := setupTestDB(t)

	mem := createTestMemory(storage, "Test memory")

	// Mark as stale first
	storage.MarkStale(mem.ID, true)

	// Mark as verified
	err := storage.MarkVerified(mem.ID)
	if err != nil {
		t.Fatalf("Failed to mark verified: %v", err)
	}

	// Retrieve and verify
	retrieved, err := storage.Get(mem.ID)
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
	storage := setupTestDB(t)

	projectID := "test-project"

	// Create mix of stale and fresh memories
	mem1 := createTestMemory(storage, "Stale memory 1")
	mem2 := createTestMemory(storage, "Fresh memory")
	mem3 := createTestMemory(storage, "Stale memory 2")

	storage.MarkStale(mem1.ID, true)
	storage.MarkStale(mem3.ID, true)

	// Get stale memories
	stale, err := storage.GetStaleMemories(projectID)
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
	storage := setupTestDB(t)

	mem := createTestMemory(storage, "Original content")

	// Update content
	mem.Content = "Updated content"
	mem.UpdatedAt = time.Now()

	err := storage.Update(mem)
	if err != nil {
		t.Fatalf("Failed to update memory: %v", err)
	}

	// Retrieve and verify
	retrieved, err := storage.Get(mem.ID)
	if err != nil {
		t.Fatalf("Failed to get memory: %v", err)
	}

	if retrieved.Content != "Updated content" {
		t.Errorf("Content not updated: got %q, want %q", retrieved.Content, "Updated content")
	}
}

func TestDelete(t *testing.T) {
	storage := setupTestDB(t)

	mem := createTestMemory(storage, "To be deleted")

	// Delete memory
	err := storage.Delete(mem.ID)
	if err != nil {
		t.Fatalf("Failed to delete memory: %v", err)
	}

	// Verify it's gone
	retrieved, err := storage.Get(mem.ID)
	if err == nil || err.Error() != "memory not found" {
		t.Fatalf("Expected 'memory not found' error, got: %v", err)
	}

	if retrieved != nil {
		t.Error("Memory should be deleted")
	}
}

func TestSearch(t *testing.T) {
	storage := setupTestDB(t)

	// Create memories with searchable content
	createMemoryWithContent(storage, "Authentication bug fix with JWT tokens")
	createMemoryWithContent(storage, "Cache optimization for session storage")
	createMemoryWithContent(storage, "JWT token validation improvements")

	// Search for "JWT"
	query := memory.SearchQuery{
		Query:     "JWT",
		ProjectID: "test-project",
		Limit:     10,
	}

	results, err := storage.Search(query)
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 results for 'JWT', got %d", len(results))
	}
}

func TestList(t *testing.T) {
	storage := setupTestDB(t)

	// Create several memories
	for i := 0; i < 5; i++ {
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
		createTestMemory(storage, "Memory "+string(rune('A'+i)))
	}

	// List with limit
	memories, err := storage.List("test-project", 3, nil)
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

func createTestMemory(storage *SQLiteStorage, content string) *memory.Memory {
	mem := &memory.Memory{
		ID:        uuid.New().String(),
		ProjectID: "test-project",
		Content:   content,
		Type:      memory.TypeGeneral,
		Priority:  "normal",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	storage.Create(mem)
	return mem
}

func createMemoryWithContent(storage *SQLiteStorage, content string) *memory.Memory {
	return createTestMemory(storage, content)
}
