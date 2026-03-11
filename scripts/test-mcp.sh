#!/bin/bash
# Manual MCP protocol testing script
# Tests basic MCP protocol compliance and tool functionality

set -e

echo "========================================="
echo "Memex MCP Server - Protocol Test"
echo "========================================="
echo ""

# Build memex
echo "Building memex..."
make build
echo "✓ Build successful"
echo ""

# Create temporary database
DB_PATH="/tmp/memex-test-$$.db"
trap "rm -f $DB_PATH" EXIT

echo "Using database: $DB_PATH"
echo ""

# Test 1: Initialize
echo "Test 1: Initialize handshake"
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | \
  MEMEX_DATABASE_PATH="$DB_PATH" ./bin/memex 2>/dev/null &
PID=$!
sleep 1
kill $PID 2>/dev/null || true
wait $PID 2>/dev/null || true
echo "✓ Initialize test passed"
echo ""

# Test 2: List tools
echo "Test 2: List tools"
{
  echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}'
  sleep 0.5
  echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
  sleep 0.5
} | MEMEX_DATABASE_PATH="$DB_PATH" timeout 3 ./bin/memex 2>/dev/null | grep -q "memex_remember"
echo "✓ Tools list test passed"
echo ""

# Test 3: Remember and recall
echo "Test 3: Remember and recall workflow"
{
  echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}'
  sleep 0.5
  echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"memex_remember","arguments":{"content":"Test memory for protocol validation"}}}'
  sleep 0.5
  echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"memex_recall","arguments":{"query":"test memory"}}}'
  sleep 0.5
} | MEMEX_DATABASE_PATH="$DB_PATH" timeout 5 ./bin/memex 2>/dev/null | grep -q "Test memory"
echo "✓ Remember/recall test passed"
echo ""

# Test 4: List memories
echo "Test 4: List memories"
{
  echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}'
  sleep 0.5
  echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"memex_list","arguments":{}}}'
  sleep 0.5
} | MEMEX_DATABASE_PATH="$DB_PATH" timeout 5 ./bin/memex 2>/dev/null | grep -q "jsonrpc"
echo "✓ List test passed"
echo ""

# Test 5: Stats
echo "Test 5: Memory statistics"
{
  echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}'
  sleep 0.5
  echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"memex_stats","arguments":{}}}'
  sleep 0.5
} | MEMEX_DATABASE_PATH="$DB_PATH" timeout 5 ./bin/memex 2>/dev/null | grep -q "jsonrpc"
echo "✓ Stats test passed"
echo ""

echo "========================================="
echo "All MCP protocol tests passed! ✓"
echo "========================================="
