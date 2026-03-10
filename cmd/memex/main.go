package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/AnkushinDaniil/memex/internal/config"
	"github.com/AnkushinDaniil/memex/internal/mcp"
	"github.com/AnkushinDaniil/memex/internal/memory"
	"github.com/AnkushinDaniil/memex/internal/storage"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Create SQLite storage
	store, err := storage.NewSQLite(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	// Initialize storage (create tables)
	ctx := context.Background()
	if err := store.Initialize(ctx); err != nil {
		// Close storage before fatal exit
		if closeErr := store.Close(); closeErr != nil {
			log.Printf("Error closing storage: %v", closeErr)
		}
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Create memory service
	memoryService := memory.NewService(store)

	// Create MCP server
	mcpServer := mcp.NewServer(memoryService)

	// Log startup (to stderr so it doesn't interfere with stdio protocol)
	fmt.Fprintf(os.Stderr, "Memex MCP Server starting...\n")
	fmt.Fprintf(os.Stderr, "Database: %s\n", cfg.DatabasePath)
	fmt.Fprintf(os.Stderr, "Mode: %s\n", cfg.Mode)
	fmt.Fprintf(os.Stderr, "Listening on stdin/stdout...\n")

	// Start serving via stdio
	if err := mcpServer.ServeStdio(); err != nil {
		// Close storage before fatal exit
		if closeErr := store.Close(); closeErr != nil {
			log.Printf("Error closing storage: %v", closeErr)
		}
		log.Fatalf("Server error: %v", err)
	}

	// Close storage on normal exit
	if closeErr := store.Close(); closeErr != nil {
		log.Printf("Error closing storage: %v", closeErr)
	}
}
