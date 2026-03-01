#!/bin/bash

echo "Test 1: Model from environment variable"
RALPH_MODEL=gpt-4 DEBUG=1 ./ralph --agent opencode 2>&1 | grep -A 2 "Using agent"

echo -e "\nTest 2: Model from command-line flag"
DEBUG=1 ./ralph --agent claude --model claude-opus 2>&1 | grep -A 2 "Using agent"

echo -e "\nTest 3: Model from config file"
DEBUG=1 ./ralph --config test_config.ralphrc 2>&1 | grep "Using agent"
