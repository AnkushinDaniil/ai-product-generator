package mcp

// GetTools returns all available MCP tools
func GetTools() []Tool {
	return []Tool{
		{
			Name:        "memex_remember",
			Description: "Store a new memory with optional code anchors, tags, and metadata",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"content": map[string]interface{}{
						"type":        "string",
						"description": "The content to remember",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]string{"type": "string"},
						"description": "Tags for categorization (optional)",
					},
					"priority": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"low", "normal", "high"},
						"default":     "normal",
						"description": "Priority level (optional)",
					},
					"type": map[string]interface{}{
						"type": "string",
						"enum": []string{
							"bug-fix", "gotcha", "connection", "design-decision",
							"aha-moment", "refactoring", "performance", "security", "general",
						},
						"default":     "general",
						"description": "Type of memory (optional)",
					},
					"anchors": map[string]interface{}{
						"type":        "array",
						"description": "Code anchors for location tracking (optional)",
						"items": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"file": map[string]interface{}{
									"type":        "string",
									"description": "File path relative to project root",
								},
								"function": map[string]interface{}{
									"type":        "string",
									"description": "Function/method name (optional)",
								},
								"start_line": map[string]interface{}{
									"type":        "integer",
									"description": "Start line number (optional)",
								},
								"end_line": map[string]interface{}{
									"type":        "integer",
									"description": "End line number (optional)",
								},
							},
							"required": []string{"file"},
						},
					},
				},
				Required: []string{"content"},
			},
		},
		{
			Name:        "memex_recall",
			Description: "Search memories using full-text search with BM25 ranking",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query for full-text search",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"default":     10,
						"description": "Maximum number of results (optional)",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]string{"type": "string"},
						"description": "Filter by tags (optional)",
					},
					"type": map[string]interface{}{
						"type": "string",
						"enum": []string{
							"bug-fix", "gotcha", "connection", "design-decision",
							"aha-moment", "refactoring", "performance", "security", "general",
						},
						"description": "Filter by memory type (optional)",
					},
				},
				Required: []string{"query"},
			},
		},
		{
			Name:        "memex_forget",
			Description: "Delete a memory by ID",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"memory_id": map[string]interface{}{
						"type":        "string",
						"description": "ID of the memory to delete",
					},
				},
				Required: []string{"memory_id"},
			},
		},
		{
			Name:        "memex_list",
			Description: "List recent memories for the current project",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "integer",
						"default":     20,
						"description": "Maximum number of results (optional)",
					},
					"tags": map[string]interface{}{
						"type":        "array",
						"items":       map[string]string{"type": "string"},
						"description": "Filter by tags (optional)",
					},
				},
				Required: []string{},
			},
		},
		{
			Name:        "memex_stats",
			Description: "Get statistics about stored memories",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"project_id": map[string]interface{}{
						"type":        "string",
						"description": "Project ID to get stats for (optional, defaults to current project)",
					},
				},
				Required: []string{},
			},
		},
	}
}
