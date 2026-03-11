package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/AnkushinDaniil/memex/internal/memory"
)

const maxContentLength = 1_000_000 // 1MB

// executeRemember implements the memex_remember tool
func (s *Server) executeRemember(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract required content
	content, ok := params["content"].(string)
	if !ok || content == "" {
		return nil, fmt.Errorf("content is required")
	}

	// Validate content length
	if len(content) > maxContentLength {
		return nil, fmt.Errorf("content exceeds maximum length of %d bytes", maxContentLength)
	}

	// Extract optional parameters with validation
	tags := extractTags(params)
	priority := extractPriority(params)
	memType := extractMemoryType(params)
	anchors := extractAnchors(params)

	// Use cached project ID
	projectID := s.projectID

	// Create memory
	mem, err := s.memoryService.Remember(ctx, content, tags, priority, projectID, memType, anchors)
	if err != nil {
		return nil, fmt.Errorf("failed to store memory: %w", err)
	}

	return map[string]interface{}{
		"memory_id":  mem.ID,
		"created_at": mem.CreatedAt.Format(time.RFC3339),
		"project_id": mem.ProjectID,
	}, nil
}

// executeRecall implements the memex_recall tool
func (s *Server) executeRecall(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract required query
	queryStr, ok := params["query"].(string)
	if !ok || queryStr == "" {
		return nil, fmt.Errorf("query is required")
	}

	// Build search query
	searchQuery := &memory.SearchQuery{
		Query:     queryStr,
		ProjectID: s.projectID,
		Limit:     extractLimit(params, 10),
		Tags:      extractTags(params),
		Type:      extractMemoryType(params),
	}

	// Search memories
	results, err := s.memoryService.Recall(ctx, searchQuery)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Format results
	memories := make([]map[string]interface{}, len(results))
	for i, result := range results {
		memories[i] = map[string]interface{}{
			"memory_id":  result.Memory.ID,
			"content":    result.Memory.Content,
			"type":       result.Memory.Type,
			"tags":       result.Memory.Tags,
			"priority":   result.Memory.Priority,
			"score":      result.Score,
			"is_stale":   result.Memory.IsStale,
			"created_at": result.Memory.CreatedAt.Format(time.RFC3339),
		}
	}

	return map[string]interface{}{
		"memories": memories,
		"count":    len(memories),
	}, nil
}

// executeForget implements the memex_forget tool
func (s *Server) executeForget(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract required memory_id
	memoryID, ok := params["memory_id"].(string)
	if !ok || memoryID == "" {
		return nil, fmt.Errorf("memory_id is required")
	}

	// Delete memory
	err := s.memoryService.Forget(ctx, memoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete memory: %w", err)
	}

	return map[string]interface{}{
		"success": true,
		"message": "Memory deleted successfully",
	}, nil
}

// executeList implements the memex_list tool
func (s *Server) executeList(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	projectID := s.projectID
	limit := extractLimit(params, 20)
	tags := extractTags(params)

	// List memories
	memories, err := s.memoryService.List(ctx, projectID, limit, tags)
	if err != nil {
		return nil, fmt.Errorf("failed to list memories: %w", err)
	}

	// Format results
	result := make([]map[string]interface{}, len(memories))
	for i, mem := range memories {
		result[i] = map[string]interface{}{
			"memory_id":  mem.ID,
			"content":    mem.Content,
			"type":       mem.Type,
			"tags":       mem.Tags,
			"priority":   mem.Priority,
			"is_stale":   mem.IsStale,
			"created_at": mem.CreatedAt.Format(time.RFC3339),
		}
	}

	return map[string]interface{}{
		"memories": result,
		"count":    len(result),
	}, nil
}

// executeStats implements the memex_stats tool
func (s *Server) executeStats(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	projectID := s.projectID
	if pid, ok := params["project_id"].(string); ok && pid != "" {
		projectID = pid
	}

	// List all memories to count
	memories, err := s.memoryService.List(ctx, projectID, 10000, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	// Count by type
	typeCount := make(map[string]int)
	staleCount := 0
	for i := range memories {
		typeCount[string(memories[i].Type)]++
		if memories[i].IsStale {
			staleCount++
		}
	}

	return map[string]interface{}{
		"total_memories": len(memories),
		"stale_memories": staleCount,
		"by_type":        typeCount,
		"project_id":     projectID,
	}, nil
}

// Helper functions with validation

func extractTags(params map[string]interface{}) []string {
	if tagsRaw, ok := params["tags"]; ok {
		if tagsSlice, ok := tagsRaw.([]interface{}); ok {
			tags := make([]string, 0, len(tagsSlice))
			for _, t := range tagsSlice {
				if tagStr, ok := t.(string); ok {
					tags = append(tags, tagStr)
				}
			}
			return tags
		}
	}
	return nil
}

func extractPriority(params map[string]interface{}) string {
	if priority, ok := params["priority"].(string); ok {
		switch priority {
		case "low", "normal", "high":
			return priority
		}
	}
	return "normal"
}

func extractMemoryType(params map[string]interface{}) memory.MemoryType {
	if memType, ok := params["type"].(string); ok {
		switch memory.MemoryType(memType) {
		case memory.TypeBugFix, memory.TypeGotcha, memory.TypeConnection,
			memory.TypeDesignDecision, memory.TypeAhaMoment, memory.TypeRefactoring,
			memory.TypePerformance, memory.TypeSecurity, memory.TypeGeneral:
			return memory.MemoryType(memType)
		}
	}
	return memory.TypeGeneral
}

func extractLimit(params map[string]interface{}, defaultLimit int) int {
	if limit, ok := params["limit"].(float64); ok {
		if limit > 0 && limit <= 1000 {
			return int(limit)
		}
	}
	return defaultLimit
}

func extractAnchors(params map[string]interface{}) []memory.CodeAnchor {
	if anchorsRaw, ok := params["anchors"]; ok {
		if anchorsSlice, ok := anchorsRaw.([]interface{}); ok {
			anchors := make([]memory.CodeAnchor, 0, len(anchorsSlice))
			for _, a := range anchorsSlice {
				if anchorMap, ok := a.(map[string]interface{}); ok {
					anchor := memory.CodeAnchor{}

					// Validate file is non-empty
					if file, ok := anchorMap["file"].(string); ok && file != "" {
						anchor.File = file
					} else {
						continue // Skip invalid anchors
					}

					if function, ok := anchorMap["function"].(string); ok {
						anchor.Function = function
					}

					var startLine, endLine int
					if sl, ok := anchorMap["start_line"].(float64); ok {
						startLine = int(sl)
					}
					if el, ok := anchorMap["end_line"].(float64); ok {
						endLine = int(el)
					}

					// Validate line numbers
					if startLine > 0 && endLine > 0 {
						if startLine <= endLine {
							anchor.StartLine = startLine
							anchor.EndLine = endLine
						} else {
							continue // Skip invalid line ranges
						}
					} else if startLine > 0 || endLine > 0 {
						continue // Skip partial line ranges
					}

					anchors = append(anchors, anchor)
				}
			}
			return anchors
		}
	}
	return nil
}
