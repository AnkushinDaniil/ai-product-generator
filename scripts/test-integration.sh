#!/bin/bash
# Integration test script
# Tests full end-to-end functionality with realistic scenarios

set -e

echo "========================================="
echo "Memex MCP Server - Integration Test"
echo "========================================="
echo ""

# Build memex
echo "Building memex..."
make build
echo "✓ Build successful"
echo ""

# Create temporary database
DB_PATH="/tmp/memex-integration-test-$$.db"
trap "rm -f $DB_PATH" EXIT

echo "Using database: $DB_PATH"
echo ""

# Helper function to send MCP commands
send_mcp() {
  local commands="$1"
  echo "$commands" | MEMEX_DATABASE_PATH="$DB_PATH" timeout 10 ./bin/memex 2>/dev/null
}

# Test 1: Server starts and initializes
echo "Test 1: Server initialization"
RESPONSE=$(send_mcp '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}')
if echo "$RESPONSE" | grep -q "2024-11-05"; then
  echo "✓ Server initialized with protocol version 2024-11-05"
else
  echo "✗ Initialization failed"
  exit 1
fi
echo ""

# Test 2: Tools are listed correctly
echo "Test 2: Tool discovery"
COMMANDS=$(cat <<EOF
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}
{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}
EOF
)
RESPONSE=$(send_mcp "$COMMANDS")
TOOL_COUNT=$(echo "$RESPONSE" | grep -o "memex_" | wc -l)
if [ "$TOOL_COUNT" -ge 5 ]; then
  echo "✓ Found $TOOL_COUNT MCP tools"
else
  echo "✗ Expected 5 tools, found $TOOL_COUNT"
  exit 1
fi
echo ""

# Test 3: Store memory with tags and priority
echo "Test 3: Store memory with metadata"
COMMANDS=$(cat <<EOF
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"memex_remember","arguments":{"content":"We use PostgreSQL for the database","tags":["database","postgres"],"priority":"high","type":"design-decision"}}}
EOF
)
RESPONSE=$(send_mcp "$COMMANDS")
if echo "$RESPONSE" | grep -q "memory_id"; then
  echo "✓ Memory stored successfully with metadata"
else
  echo "✗ Failed to store memory"
  exit 1
fi
echo ""

# Test 4: Search retrieves correct memory
echo "Test 4: Full-text search"
COMMANDS=$(cat <<EOF
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"memex_recall","arguments":{"query":"PostgreSQL database"}}}
EOF
)
RESPONSE=$(send_mcp "$COMMANDS")
if echo "$RESPONSE" | grep -q "PostgreSQL"; then
  echo "✓ Search returned correct memory"
else
  echo "✗ Search failed to find memory"
  exit 1
fi
echo ""

# Test 5: List memories
echo "Test 5: List recent memories"
COMMANDS=$(cat <<EOF
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"memex_list","arguments":{"limit":10}}}
EOF
)
RESPONSE=$(send_mcp "$COMMANDS")
if echo "$RESPONSE" | grep -q "PostgreSQL"; then
  echo "✓ List returned stored memories"
else
  echo "✗ List failed"
  exit 1
fi
echo ""

# Test 6: Store memory with code anchor
echo "Test 6: Store memory with code anchor"
COMMANDS=$(cat <<EOF
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"memex_remember","arguments":{"content":"JWT token expiry is 24 hours","anchors":[{"file":"internal/auth/jwt.go","function":"GenerateToken","start_line":45,"end_line":60}]}}}
EOF
)
RESPONSE=$(send_mcp "$COMMANDS")
if echo "$RESPONSE" | grep -q "memory_id"; then
  echo "✓ Memory with code anchor stored successfully"
else
  echo "✗ Failed to store memory with anchor"
  exit 1
fi
echo ""

# Test 7: Get statistics
echo "Test 7: Memory statistics"
COMMANDS=$(cat <<EOF
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"memex_stats","arguments":{}}}
EOF
)
RESPONSE=$(send_mcp "$COMMANDS")
if echo "$RESPONSE" | grep -q "total_memories"; then
  echo "✓ Statistics retrieved successfully"
else
  echo "✗ Failed to get statistics"
  exit 1
fi
echo ""

# Test 8: Persistence across sessions
echo "Test 8: Persistence across sessions"
COMMANDS=$(cat <<EOF
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}
{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"memex_recall","arguments":{"query":"PostgreSQL"}}}
EOF
)
RESPONSE=$(send_mcp "$COMMANDS")
if echo "$RESPONSE" | grep -q "PostgreSQL"; then
  echo "✓ Memories persist across sessions"
else
  echo "✗ Persistence test failed"
  exit 1
fi
echo ""

echo "========================================="
echo "All integration tests passed! ✓"
echo "========================================="
echo ""
echo "Summary:"
echo "  - Server initialization: ✓"
echo "  - Tool discovery: ✓"
echo "  - Store with metadata: ✓"
echo "  - Full-text search: ✓"
echo "  - List memories: ✓"
echo "  - Code anchors: ✓"
echo "  - Statistics: ✓"
echo "  - Persistence: ✓"
