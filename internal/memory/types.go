package memory

import "time"

// Memory represents a stored memory with code context
type Memory struct {
	// Identity
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`

	// Content
	Content string     `json:"content"`
	Type    MemoryType `json:"type"`

	// Code Location
	Anchors []CodeAnchor `json:"anchors,omitempty"`

	// Relationships
	ConnectedTo []string `json:"connected_to,omitempty"` // Memory IDs

	// Metadata
	Tags     []string `json:"tags,omitempty"`
	Priority string   `json:"priority"`

	// Change Tracking
	CodeHash     string     `json:"code_hash,omitempty"`
	IsStale      bool       `json:"is_stale"`
	LastVerified *time.Time `json:"last_verified,omitempty"`

	// Usage
	RetrievalCount int       `json:"retrieval_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
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
	MemoryID  string `json:"memory_id,omitempty"` // Reference to parent memory
	File      string `json:"file"`                // Relative path
	Function  string `json:"function,omitempty"`
	StartLine int    `json:"start_line,omitempty"`
	EndLine   int    `json:"end_line,omitempty"`
	GitCommit string `json:"git_commit,omitempty"`
}

// MemoryConnection represents a relationship between memories
type MemoryConnection struct {
	ID           string         `json:"id"`
	FromMemoryID string         `json:"from_memory_id"`
	ToMemoryID   string         `json:"to_memory_id"`
	Relationship ConnectionType `json:"relationship"`
	Description  string         `json:"description,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
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
	StartLine int
	EndLine   int
	Code      string // Actual code content for hashing
}

// MemoryWithRelevance includes relevance scoring
type MemoryWithRelevance struct {
	Memory    Memory  `json:"memory"`
	Relevance float64 `json:"relevance"`
	Reason    string  `json:"reason"`
}

// SearchQuery represents parameters for searching memories
type SearchQuery struct {
	Query     string
	ProjectID string
	Tags      []string
	Type      MemoryType
	Since     *time.Time
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
	Initialize() error

	// Close closes the storage connection
	Close() error

	// Memory CRUD
	Create(mem *Memory) error
	Get(id string) (*Memory, error)
	Update(mem *Memory) error
	Delete(id string) error

	// Search and retrieval
	Search(query SearchQuery) ([]SearchResult, error)
	List(projectID string, limit int, tags []string) ([]Memory, error)

	// Code anchors
	CreateAnchor(anchor *CodeAnchor) error
	GetAnchorsByMemory(memoryID string) ([]CodeAnchor, error)
	FindMemoriesByAnchor(file string, line int) ([]Memory, error)
	FindMemoriesInFile(file string) ([]Memory, error)

	// Memory connections
	CreateConnection(conn *MemoryConnection) error
	GetConnections(memoryID string) ([]MemoryConnection, error)
	GetConnectedMemories(memoryID string, depth int) ([]Memory, error)

	// Staleness tracking
	MarkStale(memoryID string, isStale bool) error
	MarkVerified(memoryID string) error
	GetStaleMemories(projectID string) ([]Memory, error)
}
