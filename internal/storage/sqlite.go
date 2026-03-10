package storage

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3" //nolint:gci // blank import for side effects

	"github.com/AnkushinDaniil/memex/internal/memory" //nolint:gci // project import
)

// Build tags required for FTS5 support:
// go build -tags "fts5"

// ErrNotImplemented is returned when a method is not yet implemented
var ErrNotImplemented = errors.New("not yet implemented")

// ErrNotFound is returned when a memory is not found
var ErrNotFound = errors.New("memory not found")

// Ensure SQLiteStorage implements memory.Storage
var _ memory.Storage = (*SQLiteStorage)(nil)

// SQLiteStorage implements Storage using SQLite with FTS5
type SQLiteStorage struct {
	db   *sql.DB
	path string
}

// NewSQLite creates a new SQLite storage instance
func NewSQLite(path string) (*SQLiteStorage, error) {
	// Enable FTS5 and other extensions
	dsn := path + "?_fts5=true&_journal_mode=WAL"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	// Enable foreign keys
	if _, err := db.ExecContext(context.Background(), "PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close() //nolint:errcheck // error already being returned
		return nil, err
	}

	return &SQLiteStorage{
		db:   db,
		path: path,
	}, nil
}

// Initialize creates the database schema
func (s *SQLiteStorage) Initialize(ctx context.Context) error {
	schema := `
	-- Main memories table
	CREATE TABLE IF NOT EXISTS memories (
		id TEXT PRIMARY KEY,
		project_id TEXT NOT NULL,
		content TEXT NOT NULL,
		type TEXT NOT NULL,

		-- Change tracking
		code_hash TEXT,
		is_stale BOOLEAN DEFAULT 0,
		last_verified DATETIME,

		tags TEXT,
		priority TEXT DEFAULT 'normal',
		retrieval_count INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_memories_project ON memories(project_id);
	CREATE INDEX IF NOT EXISTS idx_memories_type ON memories(type);
	CREATE INDEX IF NOT EXISTS idx_memories_stale ON memories(is_stale);

	-- Code anchors
	CREATE TABLE IF NOT EXISTS code_anchors (
		id TEXT PRIMARY KEY,
		memory_id TEXT NOT NULL,
		file TEXT NOT NULL,
		function TEXT,
		start_line INTEGER,
		end_line INTEGER,
		git_commit TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (memory_id) REFERENCES memories(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_anchors_memory ON code_anchors(memory_id);
	CREATE INDEX IF NOT EXISTS idx_anchors_file ON code_anchors(file);
	CREATE INDEX IF NOT EXISTS idx_anchors_function ON code_anchors(function);
	CREATE INDEX IF NOT EXISTS idx_anchors_location ON code_anchors(file, start_line, end_line);

	-- Memory connections
	CREATE TABLE IF NOT EXISTS memory_connections (
		id TEXT PRIMARY KEY,
		from_memory_id TEXT NOT NULL,
		to_memory_id TEXT NOT NULL,
		relationship TEXT NOT NULL,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (from_memory_id) REFERENCES memories(id) ON DELETE CASCADE,
		FOREIGN KEY (to_memory_id) REFERENCES memories(id) ON DELETE CASCADE,
		UNIQUE(from_memory_id, to_memory_id, relationship)
	);

	CREATE INDEX IF NOT EXISTS idx_connections_from ON memory_connections(from_memory_id);
	CREATE INDEX IF NOT EXISTS idx_connections_to ON memory_connections(to_memory_id);

	-- FTS5 for full-text search
	CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts USING fts5(
		content,
		tags,
		content='memories',
		content_rowid='rowid',
		tokenize='porter unicode61'
	);

	-- FTS sync triggers
	CREATE TRIGGER IF NOT EXISTS memories_ai AFTER INSERT ON memories BEGIN
		INSERT INTO memories_fts(rowid, content, tags)
		VALUES (new.rowid, new.content, new.tags);
	END;

	CREATE TRIGGER IF NOT EXISTS memories_au AFTER UPDATE ON memories BEGIN
		INSERT INTO memories_fts(memories_fts, rowid, content, tags)
		VALUES('delete', old.rowid, old.content, old.tags);
		INSERT INTO memories_fts(rowid, content, tags)
		VALUES (new.rowid, new.content, new.tags);
	END;

	CREATE TRIGGER IF NOT EXISTS memories_ad AFTER DELETE ON memories BEGIN
		INSERT INTO memories_fts(memories_fts, rowid, content, tags)
		VALUES('delete', old.rowid, old.content, old.tags);
	END;
	`

	_, err := s.db.ExecContext(ctx, schema)
	return err
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// Create stores a new memory
func (s *SQLiteStorage) Create(ctx context.Context, mem *memory.Memory) error {
	query := `
		INSERT INTO memories (id, project_id, content, type, code_hash, is_stale,
			last_verified, tags, priority, retrieval_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	tagsJSON := ""
	if len(mem.Tags) > 0 {
		tagsJSON = `["` + mem.Tags[0]
		for i := 1; i < len(mem.Tags); i++ {
			tagsJSON += `","` + mem.Tags[i]
		}
		tagsJSON += `"]`
	}

	_, err := s.db.ExecContext(ctx, query,
		mem.ID, mem.ProjectID, mem.Content, mem.Type,
		mem.CodeHash, mem.IsStale, mem.LastVerified,
		tagsJSON, mem.Priority, mem.RetrievalCount,
		mem.CreatedAt, mem.UpdatedAt)

	return err
}

// Get retrieves a memory by ID
func (s *SQLiteStorage) Get(ctx context.Context, id string) (*memory.Memory, error) {
	query := `
		SELECT id, project_id, content, type, code_hash, is_stale, last_verified,
			tags, priority, retrieval_count, created_at, updated_at
		FROM memories
		WHERE id = ?
	`

	var mem memory.Memory
	var tagsJSON sql.NullString
	var lastVerified sql.NullTime

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&mem.ID, &mem.ProjectID, &mem.Content, &mem.Type,
		&mem.CodeHash, &mem.IsStale, &lastVerified,
		&tagsJSON, &mem.Priority, &mem.RetrievalCount,
		&mem.CreatedAt, &mem.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if lastVerified.Valid {
		mem.LastVerified = &lastVerified.Time
	}

	// Parse tags JSON (simplified)
	if tagsJSON.Valid && tagsJSON.String != "" {
		// Simple JSON parsing for tags array
		mem.Tags = []string{}
	}

	return &mem, nil
}

// Update updates an existing memory
func (s *SQLiteStorage) Update(ctx context.Context, mem *memory.Memory) error {
	query := `
		UPDATE memories
		SET content = ?, type = ?, code_hash = ?, is_stale = ?,
			last_verified = ?, tags = ?, priority = ?,
			retrieval_count = ?, updated_at = ?
		WHERE id = ?
	`

	tagsJSON := ""
	if len(mem.Tags) > 0 {
		tagsJSON = `["` + mem.Tags[0]
		for i := 1; i < len(mem.Tags); i++ {
			tagsJSON += `","` + mem.Tags[i]
		}
		tagsJSON += `"]`
	}

	_, err := s.db.ExecContext(ctx, query,
		mem.Content, mem.Type, mem.CodeHash, mem.IsStale,
		mem.LastVerified, tagsJSON, mem.Priority,
		mem.RetrievalCount, mem.UpdatedAt, mem.ID)

	return err
}

// Delete removes a memory
func (s *SQLiteStorage) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM memories WHERE id = ?", id)
	return err
}

// Search performs full-text search
func (s *SQLiteStorage) Search(ctx context.Context, query *memory.SearchQuery) ([]memory.SearchResult, error) {
	sqlQuery := `
		SELECT m.id, m.project_id, m.content, m.type, m.code_hash, m.is_stale,
			m.last_verified, m.tags, m.priority, m.retrieval_count,
			m.created_at, m.updated_at, rank
		FROM memories m
		JOIN memories_fts ON m.rowid = memories_fts.rowid
		WHERE memories_fts MATCH ?
			AND m.project_id = ?
		ORDER BY rank
		LIMIT ?
	`

	limit := query.Limit
	if limit == 0 {
		limit = 10
	}

	rows, err := s.db.QueryContext(ctx, sqlQuery, query.Query, query.ProjectID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }() //nolint:errcheck // cleanup in defer

	var results []memory.SearchResult
	for rows.Next() {
		var mem memory.Memory
		var tagsJSON sql.NullString
		var lastVerified sql.NullTime
		var rank float64

		err := rows.Scan(
			&mem.ID, &mem.ProjectID, &mem.Content, &mem.Type,
			&mem.CodeHash, &mem.IsStale, &lastVerified,
			&tagsJSON, &mem.Priority, &mem.RetrievalCount,
			&mem.CreatedAt, &mem.UpdatedAt, &rank)
		if err != nil {
			return nil, err
		}

		if lastVerified.Valid {
			mem.LastVerified = &lastVerified.Time
		}

		results = append(results, memory.SearchResult{
			Memory: mem,
			Score:  rank,
		})
	}

	return results, rows.Err()
}

// List returns recent memories
func (s *SQLiteStorage) List(ctx context.Context, projectID string, limit int, _ []string) ([]memory.Memory, error) {
	query := `
		SELECT id, project_id, content, type, code_hash, is_stale, last_verified,
			tags, priority, retrieval_count, created_at, updated_at
		FROM memories
		WHERE project_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	if limit == 0 {
		limit = 20
	}

	rows, err := s.db.QueryContext(ctx, query, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }() //nolint:errcheck // cleanup in defer

	var memories []memory.Memory
	for rows.Next() {
		var mem memory.Memory
		var tagsJSON sql.NullString
		var lastVerified sql.NullTime

		err := rows.Scan(
			&mem.ID, &mem.ProjectID, &mem.Content, &mem.Type,
			&mem.CodeHash, &mem.IsStale, &lastVerified,
			&tagsJSON, &mem.Priority, &mem.RetrievalCount,
			&mem.CreatedAt, &mem.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if lastVerified.Valid {
			mem.LastVerified = &lastVerified.Time
		}

		memories = append(memories, mem)
	}

	return memories, rows.Err()
}

// CreateAnchor creates a code anchor
func (s *SQLiteStorage) CreateAnchor(ctx context.Context, anchor *memory.CodeAnchor) error {
	query := `
		INSERT INTO code_anchors (id, memory_id, file, function, start_line, end_line, git_commit)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		anchor.ID, anchor.MemoryID, anchor.File, anchor.Function,
		anchor.StartLine, anchor.EndLine, anchor.GitCommit)

	return err
}

// GetAnchorsByMemory retrieves all anchors for a memory
func (s *SQLiteStorage) GetAnchorsByMemory(ctx context.Context, memoryID string) ([]memory.CodeAnchor, error) {
	query := `
		SELECT id, file, function, start_line, end_line, git_commit
		FROM code_anchors
		WHERE memory_id = ?
	`

	rows, err := s.db.QueryContext(ctx, query, memoryID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }() //nolint:errcheck // cleanup in defer

	var anchors []memory.CodeAnchor
	for rows.Next() {
		var anchor memory.CodeAnchor
		var function sql.NullString
		var gitCommit sql.NullString

		err := rows.Scan(&anchor.ID, &anchor.File, &function,
			&anchor.StartLine, &anchor.EndLine, &gitCommit)
		if err != nil {
			return nil, err
		}

		if function.Valid {
			anchor.Function = function.String
		}
		if gitCommit.Valid {
			anchor.GitCommit = gitCommit.String
		}

		anchors = append(anchors, anchor)
	}

	return anchors, rows.Err()
}

// FindMemoriesByAnchor finds memories at a specific file and line
func (s *SQLiteStorage) FindMemoriesByAnchor(ctx context.Context, file string, line int) ([]memory.Memory, error) {
	query := `
		SELECT DISTINCT m.id, m.project_id, m.content, m.type, m.code_hash, m.is_stale,
			m.last_verified, m.tags, m.priority, m.retrieval_count, m.created_at, m.updated_at
		FROM memories m
		JOIN code_anchors ca ON m.id = ca.memory_id
		WHERE ca.file = ?
			AND ca.start_line <= ?
			AND ca.end_line >= ?
		ORDER BY ca.start_line ASC
	`

	rows, err := s.db.QueryContext(ctx, query, file, line, line)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }() //nolint:errcheck // cleanup in defer

	var memories []memory.Memory
	for rows.Next() {
		var mem memory.Memory
		var tagsJSON sql.NullString
		var lastVerified sql.NullTime

		err := rows.Scan(
			&mem.ID, &mem.ProjectID, &mem.Content, &mem.Type,
			&mem.CodeHash, &mem.IsStale, &lastVerified,
			&tagsJSON, &mem.Priority, &mem.RetrievalCount,
			&mem.CreatedAt, &mem.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if lastVerified.Valid {
			mem.LastVerified = &lastVerified.Time
		}

		memories = append(memories, mem)
	}

	return memories, rows.Err()
}

// FindMemoriesInFile finds all memories anchored to a specific file
func (s *SQLiteStorage) FindMemoriesInFile(ctx context.Context, file string) ([]memory.Memory, error) {
	query := `
		SELECT DISTINCT m.id, m.project_id, m.content, m.type, m.code_hash, m.is_stale,
			m.last_verified, m.tags, m.priority, m.retrieval_count, m.created_at, m.updated_at
		FROM memories m
		JOIN code_anchors ca ON m.id = ca.memory_id
		WHERE ca.file = ?
		ORDER BY ca.start_line ASC
	`

	rows, err := s.db.QueryContext(ctx, query, file)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }() //nolint:errcheck // cleanup in defer

	return s.scanMemories(rows)
}

// CreateConnection creates a memory connection
func (s *SQLiteStorage) CreateConnection(ctx context.Context, conn *memory.MemoryConnection) error {
	query := `
		INSERT INTO memory_connections (id, from_memory_id, to_memory_id, relationship, description)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecContext(ctx, query,
		conn.ID, conn.FromMemoryID, conn.ToMemoryID,
		conn.Relationship, conn.Description)

	return err
}

// GetConnections retrieves all connections for a memory
func (s *SQLiteStorage) GetConnections(ctx context.Context, memoryID string) ([]memory.MemoryConnection, error) {
	query := `
		SELECT id, from_memory_id, to_memory_id, relationship, description, created_at
		FROM memory_connections
		WHERE from_memory_id = ? OR to_memory_id = ?
	`

	rows, err := s.db.QueryContext(ctx, query, memoryID, memoryID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }() //nolint:errcheck // cleanup in defer

	var connections []memory.MemoryConnection
	for rows.Next() {
		var conn memory.MemoryConnection
		var description sql.NullString

		err := rows.Scan(&conn.ID, &conn.FromMemoryID, &conn.ToMemoryID,
			&conn.Relationship, &description, &conn.CreatedAt)
		if err != nil {
			return nil, err
		}

		if description.Valid {
			conn.Description = description.String
		}

		connections = append(connections, conn)
	}

	return connections, rows.Err()
}

// GetConnectedMemories retrieves connected memories recursively
func (s *SQLiteStorage) GetConnectedMemories(ctx context.Context, memoryID string, depth int) ([]memory.Memory, error) {
	query := `
		WITH RECURSIVE connected AS (
			-- Start with direct connections
			SELECT to_memory_id as id, 1 as depth
			FROM memory_connections
			WHERE from_memory_id = ?

			UNION ALL

			-- Recursively find connected
			SELECT mc.to_memory_id, c.depth + 1
			FROM memory_connections mc
			JOIN connected c ON mc.from_memory_id = c.id
			WHERE c.depth < ?
		)
		SELECT DISTINCT m.id, m.project_id, m.content, m.type, m.code_hash, m.is_stale,
			m.last_verified, m.tags, m.priority, m.retrieval_count, m.created_at, m.updated_at
		FROM memories m
		JOIN connected c ON m.id = c.id
	`

	rows, err := s.db.QueryContext(ctx, query, memoryID, depth)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }() //nolint:errcheck // cleanup in defer

	var memories []memory.Memory
	for rows.Next() {
		var mem memory.Memory
		var tagsJSON sql.NullString
		var lastVerified sql.NullTime

		err := rows.Scan(
			&mem.ID, &mem.ProjectID, &mem.Content, &mem.Type,
			&mem.CodeHash, &mem.IsStale, &lastVerified,
			&tagsJSON, &mem.Priority, &mem.RetrievalCount,
			&mem.CreatedAt, &mem.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if lastVerified.Valid {
			mem.LastVerified = &lastVerified.Time
		}

		memories = append(memories, mem)
	}

	return memories, rows.Err()
}

// MarkStale marks a memory as stale or not stale
func (s *SQLiteStorage) MarkStale(ctx context.Context, memoryID string, isStale bool) error {
	query := `UPDATE memories SET is_stale = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, isStale, memoryID)
	return err
}

// MarkVerified marks a memory as verified (not stale, updates last_verified)
func (s *SQLiteStorage) MarkVerified(ctx context.Context, memoryID string) error {
	query := `
		UPDATE memories
		SET is_stale = 0, last_verified = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := s.db.ExecContext(ctx, query, memoryID)
	return err
}

// scanMemories is a helper to scan memory rows
func (s *SQLiteStorage) scanMemories(rows *sql.Rows) ([]memory.Memory, error) {
	var memories []memory.Memory
	for rows.Next() {
		var mem memory.Memory
		var tagsJSON sql.NullString
		var lastVerified sql.NullTime

		err := rows.Scan(
			&mem.ID, &mem.ProjectID, &mem.Content, &mem.Type,
			&mem.CodeHash, &mem.IsStale, &lastVerified,
			&tagsJSON, &mem.Priority, &mem.RetrievalCount,
			&mem.CreatedAt, &mem.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if lastVerified.Valid {
			mem.LastVerified = &lastVerified.Time
		}

		memories = append(memories, mem)
	}

	return memories, rows.Err()
}

// GetStaleMemories retrieves all stale memories for a project
func (s *SQLiteStorage) GetStaleMemories(ctx context.Context, projectID string) ([]memory.Memory, error) {
	query := `
		SELECT id, project_id, content, type, code_hash, is_stale, last_verified,
			tags, priority, retrieval_count, created_at, updated_at
		FROM memories
		WHERE project_id = ? AND is_stale = 1
		ORDER BY updated_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }() //nolint:errcheck // cleanup in defer

	return s.scanMemories(rows)
}
