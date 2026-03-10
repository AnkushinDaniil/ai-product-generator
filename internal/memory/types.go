package memory

import (
	"context"
	"time"
)

// Memory represents a stored memory with code context
type Memory struct {
	UpdatedAt      time.Time    `json:"updated_at"`
	CreatedAt      time.Time    `json:"created_at"`
	LastVerified   *time.Time   `json:"last_verified,omitempty"`
	ProjectID      string       `json:"project_id"`
	Content        string       `json:"content"`
	Type           MemoryType   `json:"type"`
	ID             string       `json:"id"`
	Priority       string       `json:"priority"`
	CodeHash       string       `json:"code_hash,omitempty"`
	Tags           []string     `json:"tags,omitempty"`
	ConnectedTo    []string     `json:"connected_to,omitempty"`
	Anchors        []CodeAnchor `json:"anchors,omitempty"`
	RetrievalCount int          `json:"retrieval_count"`
	IsStale        bool         `json:"is_stale"`
}

// MemoryType categorizes the type of knowledge stored
type MemoryType string

const (
	TypeBugFix         MemoryType = "bug-fix"
	TypeGotcha         MemoryType = "gotcha"
	TypeConnection     MemoryType = "connection"
	TypeDesignDecision MemoryType = "design-decision"
	TypeAhaMoment      MemoryType = "aha-moment"
	TypeRefactoring    MemoryType = "refactoring"
	TypePerformance    MemoryType = "performance"
	TypeSecurity       MemoryType = "security"
	TypeGeneral        MemoryType = "general"
)

// CodeAnchor represents a precise code location
type CodeAnchor struct {
	ID        string `json:"id,omitempty"`
	MemoryID  string `json:"memory_id,omitempty"`
	File      string `json:"file"`
	Function  string `json:"function,omitempty"`
	GitCommit string `json:"git_commit,omitempty"`
	StartLine int    `json:"start_line,omitempty"`
	EndLine   int    `json:"end_line,omitempty"`
}

// MemoryConnection represents a relationship between memories
type MemoryConnection struct {
	CreatedAt    time.Time      `json:"created_at"`
	ID           string         `json:"id"`
	FromMemoryID string         `json:"from_memory_id"`
	ToMemoryID   string         `json:"to_memory_id"`
	Relationship ConnectionType `json:"relationship"`
	Description  string         `json:"description,omitempty"`
}

// ConnectionType defines how memories relate
type ConnectionType string

const (
	ConnAffects    ConnectionType = "affects"
	ConnDependsOn  ConnectionType = "depends-on"
	ConnRelated    ConnectionType = "related"
	ConnCausedBy   ConnectionType = "caused-by"
	ConnSupersedes ConnectionType = "supersedes"
)

// CodeContext represents the current code being viewed
type CodeContext struct {
	File      string
	Function  string
	Code      string
	StartLine int
	EndLine   int
}

// MemoryWithRelevance includes relevance scoring
type MemoryWithRelevance struct { //nolint:govet // field order for JSON compatibility
	Relevance float64 `json:"relevance"`
	Memory    Memory  `json:"memory"`
	Reason    string  `json:"reason"`
}

// SearchQuery represents parameters for searching memories
type SearchQuery struct {
	Since     *time.Time
	Query     string
	ProjectID string
	Type      MemoryType
	Tags      []string
	Limit     int
}

// SearchResult represents a memory with relevance score
type SearchResult struct {
	Memory Memory  `json:"memory"`
	Score  float64 `json:"score"`
}

// Storage defines the interface for memory storage backends
type Storage interface {
	// Initialize sets up the storage (create tables, etc.)
	Initialize(ctx context.Context) error

	// Close closes the storage connection
	Close() error

	// Memory CRUD
	Create(ctx context.Context, mem *Memory) error
	Get(ctx context.Context, id string) (*Memory, error)
	Update(ctx context.Context, mem *Memory) error
	Delete(ctx context.Context, id string) error

	// Search and retrieval
	Search(ctx context.Context, query *SearchQuery) ([]SearchResult, error)
	List(ctx context.Context, projectID string, limit int, tags []string) ([]Memory, error)

	// Code anchors
	CreateAnchor(ctx context.Context, anchor *CodeAnchor) error
	GetAnchorsByMemory(ctx context.Context, memoryID string) ([]CodeAnchor, error)
	FindMemoriesByAnchor(ctx context.Context, file string, line int) ([]Memory, error)
	FindMemoriesInFile(ctx context.Context, file string) ([]Memory, error)

	// Memory connections
	CreateConnection(ctx context.Context, conn *MemoryConnection) error
	GetConnections(ctx context.Context, memoryID string) ([]MemoryConnection, error)
	GetConnectedMemories(ctx context.Context, memoryID string, depth int) ([]Memory, error)

	// Staleness tracking
	MarkStale(ctx context.Context, memoryID string, isStale bool) error
	MarkVerified(ctx context.Context, memoryID string) error
	GetStaleMemories(ctx context.Context, projectID string) ([]Memory, error)
}
