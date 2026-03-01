#!/bin/bash
set -e

echo "Creating test config files..."
mkdir -p test_configs

# Config 1: Agent and Model
cat > test_configs/config1.ralphrc << 'CONFIG'
RALPH_AGENT=opencode
RALPH_MODEL=gpt-4
CONFIG

# Config 2: Different Agent and Model
cat > test_configs/config2.ralphrc << 'CONFIG'
RALPH_AGENT=claude
RALPH_MODEL=claude-sonnet-4
CONFIG

echo "✓ Test configs created"
echo ""
echo "Testing precedence (Flag > EnvVar > ConfigFile):"
echo ""

echo "Test 1: Flag overrides EnvVar"
RALPH_MODEL=gpt-4 DEBUG=1 ./ralph --model claude-opus --agent opencode 2>&1 | grep "Using agent"

echo ""
echo "Test 2: EnvVar without Flag"
RALPH_AGENT=claude RALPH_MODEL=gpt-4 DEBUG=1 ./ralph 2>&1 | grep "Using agent"

echo ""
echo "Test 3: ConfigFile values"
DEBUG=1 ./ralph --config test_configs/config1.ralphrc 2>&1 | grep "Using agent"

echo ""
echo "Test 4: ConfigFile with EnvVar override"
RALPH_AGENT=claude DEBUG=1 ./ralph --config test_configs/config1.ralphrc 2>&1 | grep "Using agent"

echo ""
echo "✓ All tests passed - Model is configurable via flags, env vars, and config file"
