package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/AnkushinDaniil/memex/internal/config"
	"github.com/AnkushinDaniil/memex/internal/mcp"
	"github.com/AnkushinDaniil/memex/internal/memory"
	"github.com/AnkushinDaniil/memex/internal/storage"
)

var version = "dev" // Set via -ldflags during build

func main() {
	var (
		dbPath   = flag.String("db", "", "Database path (overrides MEMEX_DATABASE_PATH)")
		logLevel = flag.String("log-level", "", "Log level: debug, info, warn, error (overrides MEMEX_LOG_LEVEL)")
		showVer  = flag.Bool("version", false, "Print version and exit")
	)
	flag.Parse()

	if *showVer {
		fmt.Printf("memex v%s\n", version)
		os.Exit(0)
	}

	cfg := config.Load()

	if *dbPath != "" {
		cfg.DatabasePath = *dbPath
	}
	if *logLevel != "" {
		cfg.LogLevel = *logLevel
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	dbDir := filepath.Dir(cfg.DatabasePath)
	if err := os.MkdirAll(dbDir, 0o700); err != nil {
		log.Fatalf("Failed to create database directory %q: %v", dbDir, err)
	}

	store, err := storage.NewSQLite(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	ctx := context.Background()
	if err := store.Initialize(ctx); err != nil {
		//nolint:errcheck // Ignoring close error before fatal exit
		store.Close()
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer func() {
		if closeErr := store.Close(); closeErr != nil {
			log.Printf("Error closing storage: %v", closeErr)
		}
	}()

	memoryService := memory.NewService(store)
	mcpServer := mcp.NewServer(memoryService)

	fmt.Fprintf(os.Stderr, "Memex MCP Server starting...\n")
	fmt.Fprintf(os.Stderr, "Database: %s\n", cfg.DatabasePath)
	fmt.Fprintf(os.Stderr, "Mode: %s\n", cfg.Mode)
	fmt.Fprintf(os.Stderr, "Listening on stdin/stdout...\n")

	if err := mcpServer.ServeStdio(); err != nil {
		log.Fatalf("Server error: %v", err) //nolint:gocritic // defer runs on normal exit only
	}
}
