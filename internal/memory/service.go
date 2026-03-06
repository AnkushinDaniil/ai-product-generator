package memory

import (
	"time"

	"github.com/google/uuid"
)

// Service handles memory business logic
type Service struct {
	storage Storage
}

// NewService creates a new memory service
func NewService(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

// Remember stores a new memory with optional code anchors
func (s *Service) Remember(content string, tags []string, priority, projectID string, memType MemoryType, anchors []CodeAnchor) (*Memory, error) {
	now := time.Now()

	mem := &Memory{
		ID:             uuid.New().String(),
		ProjectID:      projectID,
		Content:        content,
		Type:           memType,
		Tags:           tags,
		Priority:       priority,
		IsStale:        false,
		RetrievalCount: 0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Create the memory
	if err := s.storage.Create(mem); err != nil {
		return nil, err
	}

	// Create anchors if provided
	for i := range anchors {
		anchors[i].ID = uuid.New().String()
		anchors[i].MemoryID = mem.ID
		if err := s.storage.CreateAnchor(&anchors[i]); err != nil {
			return nil, err
		}
	}

	mem.Anchors = anchors
	return mem, nil
}

// Recall searches for memories using full-text search
func (s *Service) Recall(query SearchQuery) ([]SearchResult, error) {
	return s.storage.Search(query)
}

// RecallContext retrieves memories for a specific code context
func (s *Service) RecallContext(ctx CodeContext, projectID string) ([]MemoryWithRelevance, error) {
	var results []MemoryWithRelevance

	// 1. Exact anchor matches (highest relevance)
	if ctx.File != "" && ctx.StartLine > 0 {
		exactMatches, err := s.storage.FindMemoriesByAnchor(ctx.File, ctx.StartLine)
		if err != nil {
			return nil, err
		}

		for _, mem := range exactMatches {
			results = append(results, MemoryWithRelevance{
				Memory:    mem,
				Relevance: 1.0,
				Reason:    "exact_anchor_match",
			})
		}
	}

	// 2. Same file matches (medium relevance)
	if ctx.File != "" {
		fileMatches, err := s.storage.FindMemoriesInFile(ctx.File)
		if err != nil {
			return nil, err
		}

		for _, mem := range fileMatches {
			if !containsMemory(results, mem.ID) {
				results = append(results, MemoryWithRelevance{
					Memory:    mem,
					Relevance: 0.7,
					Reason:    "same_file",
				})
			}
		}
	}

	// 3. Check staleness for all memories
	for i := range results {
		if ctx.Code != "" {
			results[i].Memory.IsStale = DetectChanges(&results[i].Memory, ctx.Code)
		}
	}

	return results, nil
}

// Forget deletes a memory
func (s *Service) Forget(memoryID string) error {
	return s.storage.Delete(memoryID)
}

// List returns recent memories
func (s *Service) List(projectID string, limit int, tags []string) ([]Memory, error) {
	return s.storage.List(projectID, limit, tags)
}

// Connect creates a connection between two memories
func (s *Service) Connect(fromID, toID string, relationship ConnectionType, description string) (*MemoryConnection, error) {
	conn := &MemoryConnection{
		ID:           uuid.New().String(),
		FromMemoryID: fromID,
		ToMemoryID:   toID,
		Relationship: relationship,
		Description:  description,
		CreatedAt:    time.Now(),
	}

	if err := s.storage.CreateConnection(conn); err != nil {
		return nil, err
	}

	return conn, nil
}

// GetConnectedMemories retrieves memories connected to the given memory
func (s *Service) GetConnectedMemories(memoryID string, depth int) ([]Memory, error) {
	if depth == 0 {
		depth = 2 // Default depth
	}
	return s.storage.GetConnectedMemories(memoryID, depth)
}

// MarkStale marks a memory as stale
func (s *Service) MarkStale(memoryID string) error {
	return s.storage.MarkStale(memoryID, true)
}

// MarkVerified marks a memory as verified (not stale)
func (s *Service) MarkVerified(memoryID string) error {
	return s.storage.MarkVerified(memoryID)
}

// GetStaleMemories retrieves all stale memories for a project
func (s *Service) GetStaleMemories(projectID string) ([]Memory, error) {
	return s.storage.GetStaleMemories(projectID)
}

// Helper function to check if a memory is already in results
func containsMemory(results []MemoryWithRelevance, memoryID string) bool {
	for _, r := range results {
		if r.Memory.ID == memoryID {
			return true
		}
	}
	return false
}
