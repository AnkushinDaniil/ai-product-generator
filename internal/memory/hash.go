package memory

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// HashCode computes a SHA256 hash of normalized code content
func HashCode(code string) string {
	// Normalize whitespace for consistent hashing
	normalized := normalizeCode(code)

	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:])
}

// normalizeCode removes extra whitespace and normalizes line endings
func normalizeCode(code string) string {
	// Replace CRLF with LF
	code = strings.ReplaceAll(code, "\r\n", "\n")

	// Split into lines
	lines := strings.Split(code, "\n")

	// Trim each line and remove empty lines
	var normalized []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}

	// Join with single newlines
	return strings.Join(normalized, "\n")
}

// DetectChanges checks if code has changed from the stored hash
func DetectChanges(memory *Memory, currentCode string) bool {
	if memory.CodeHash == "" {
		return false // No hash to compare
	}

	currentHash := HashCode(currentCode)
	return currentHash != memory.CodeHash
}

// ComputeCodeHash is a helper to compute hash for a CodeContext
func ComputeCodeHash(ctx CodeContext) string {
	return HashCode(ctx.Code)
}
