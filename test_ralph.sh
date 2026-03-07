#!/usr/bin/env bash

# Test suite for ralph.sh
# Tests flags, environment variables, and config file handling

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
RALPH="$SCRIPT_DIR/ralph.sh"

TESTS_PASSED=0
TESTS_FAILED=0

assert_output_contains() {
	local output=$1
	local expected=$2
	local test_name=$3

	if echo "$output" | grep -Fq -- "$expected"; then
		echo "✓ $test_name"
		TESTS_PASSED=$((TESTS_PASSED + 1))
	else
		echo "✗ $test_name (output does not contain: $expected)"
		TESTS_FAILED=$((TESTS_FAILED + 1))
	fi
}

assert_output_not_contains() {
	local output=$1
	local expected=$2
	local test_name=$3

	if ! echo "$output" | grep -Fq -- "$expected"; then
		echo "✓ $test_name"
		TESTS_PASSED=$((TESTS_PASSED + 1))
	else
		echo "✗ $test_name (output should not contain: $expected)"
		TESTS_FAILED=$((TESTS_FAILED + 1))
	fi
}

# ============================================================================
# Test: Help flag (short form)
# ============================================================================
test_help_short_flag() {
	local output
	output=$("$RALPH" -h 2>&1)
	assert_output_contains "$output" "Usage:" "Help flag (-h) shows usage"
}

# ============================================================================
# Test: Help flag (long form)
# ============================================================================
test_help_long_flag() {
	local output
	output=$("$RALPH" --help 2>&1)
	assert_output_contains "$output" "Usage:" "Help flag (--help) shows usage"
}

# ============================================================================
# Test: Max iterations flag (short form)
# ============================================================================
test_max_iterations_short_flag() {
	local output
	output=$(DEBUG=1 "$RALPH" -m 3 2>&1)
	assert_output_contains "$output" "Max iterations: 3" "Max iterations flag (-m) sets value"
}

# ============================================================================
# Test: Max iterations flag (long form)
# ============================================================================
test_max_iterations_long_flag() {
	local output
	output=$(DEBUG=1 "$RALPH" --max-iterations 5 2>&1)
	assert_output_contains "$output" "Max iterations: 5" "Max iterations flag (--max-iterations) sets value"
}

# ============================================================================
# Test: Specs dir flag (short form)
# ============================================================================
test_specs_dir_short_flag() {
	local output
	output=$(DEBUG=1 "$RALPH" -s custom_specs 2>&1)
	assert_output_contains "$output" "custom_specs" "Specs dir flag (-s) sets value"
}

# ============================================================================
# Test: Specs dir flag (long form)
# ============================================================================
test_specs_dir_long_flag() {
	local output
	output=$(DEBUG=1 "$RALPH" --specs-dir custom_specs 2>&1)
	assert_output_contains "$output" "custom_specs" "Specs dir flag (--specs-dir) sets value"
}

# ============================================================================
# Test: Specs index flag (short form)
# ============================================================================
test_specs_index_short_flag() {
	local output
	output=$(DEBUG=1 "$RALPH" -i INDEX.md 2>&1)
	assert_output_contains "$output" "INDEX.md" "Specs index flag (-i) sets value"
}

# ============================================================================
# Test: Specs index flag (long form)
# ============================================================================
test_specs_index_long_flag() {
	local output
	output=$(DEBUG=1 "$RALPH" --specs-index INDEX.md 2>&1)
	assert_output_contains "$output" "INDEX.md" "Specs index flag (--specs-index) sets value"
}

# ============================================================================
# Test: No specs index flag
# ============================================================================
test_no_specs_index_flag() {
	local output
	output=$(DEBUG=1 "$RALPH" --no-specs-index 2>&1)
	assert_output_not_contains "$output" "README.md" "No specs index flag disables index"
}

# ============================================================================
# Test: Implementation plan name flag (short form)
# ============================================================================
test_impl_plan_short_flag() {
	local output
	output=$(DEBUG=1 "$RALPH" -n PLAN.md 2>&1)
	assert_output_contains "$output" "PLAN.md" "Implementation plan flag (-n) sets value"
}

# ============================================================================
# Test: Implementation plan name flag (long form)
# ============================================================================
test_impl_plan_long_flag() {
	local output
	output=$(DEBUG=1 "$RALPH" --implementation-plan-name PLAN.md 2>&1)
	assert_output_contains "$output" "PLAN.md" "Implementation plan flag (--implementation-plan-name) sets value"
}

# ============================================================================
# Test: Inline prompt flag
# ============================================================================
test_inline_prompt_flag() {
	local output
	output=$(DEBUG=1 "$RALPH" --prompt "Test prompt" 2>&1)
	assert_output_contains "$output" "Test prompt" "Inline prompt flag passes custom prompt"
}

run_with_fake_opencode() {
	local cmd="$1"
	local output
	output=$(bash -c "
		set -e
		tmpdir=\"\$(mktemp -d)\"
		args_log=\"\$tmpdir/opencode-args.log\"
		trap 'rm -rf \"\$tmpdir\"' EXIT
		cat >\"\$tmpdir/opencode\" <<'EOF'
#!/bin/sh
printf '%s\\n' \"\$*\" >>\"\$RALPH_TEST_OPENCODE_ARGS_FILE\"
printf '%s\\n' '<promise>COMPLETE</promise>'
EOF
		chmod +x \"\$tmpdir/opencode\"
		RALPH_TEST_OPENCODE_ARGS_FILE=\"\$args_log\" PATH=\"\$tmpdir:\$PATH\" $cmd >/dev/null 2>&1
		cat \"\$args_log\"
	" 2>&1)
	printf '%s\\n' "$output"
}

# ============================================================================
# Test: Agent flag (short form) passes --agent to opencode
# ============================================================================
test_agent_short_flag() {
	local output
	output=$(run_with_fake_opencode "\"$RALPH\" -m 1 -a cli-agent --prompt test")
	assert_output_contains "$output" "run --agent cli-agent" "Agent flag (-a) passes --agent to opencode"
}

# ============================================================================
# Test: Agent flag (long form) passes --agent to opencode
# ============================================================================
test_agent_long_flag() {
	local output
	output=$(run_with_fake_opencode "\"$RALPH\" -m 1 --agent long-agent --prompt test")
	assert_output_contains "$output" "run --agent long-agent" "Agent flag (--agent) passes --agent to opencode"
}

# ============================================================================
# Test: Environment variable - RALPH_AGENT
# ============================================================================
test_env_agent() {
	local output
	output=$(run_with_fake_opencode "RALPH_AGENT=env-agent \"$RALPH\" -m 1 --prompt test")
	assert_output_contains "$output" "run --agent env-agent" "RALPH_AGENT env var passes --agent to opencode"
}

# ============================================================================
# Test: Config file loading - RALPH_AGENT
# ============================================================================
test_config_agent() {
	local output
	output=$(bash -c "
		set -e
		tmpdir=\"\$(mktemp -d)\"
		trap 'rm -rf \"\$tmpdir\"' EXIT
		cat >\"\$tmpdir/.ralphrc\" <<EOF
RALPH_AGENT=config-agent
EOF
		cd \"\$tmpdir\"
		$(typeset -f run_with_fake_opencode)
		run_with_fake_opencode '\"$RALPH\" -m 1 --prompt test'
	" 2>&1)
	assert_output_contains "$output" "run --agent config-agent" "Config file RALPH_AGENT passes --agent to opencode"
}

# ============================================================================
# Test: Agent precedence - flag overrides env and config
# ============================================================================
test_agent_precedence() {
	local output
	output=$(bash -c "
		set -e
		tmpdir=\"\$(mktemp -d)\"
		trap 'rm -rf \"\$tmpdir\"' EXIT
		cat >\"\$tmpdir/.ralphrc\" <<EOF
RALPH_AGENT=config-agent
EOF
		cd \"\$tmpdir\"
		$(typeset -f run_with_fake_opencode)
		RALPH_AGENT=env-agent run_with_fake_opencode '\"$RALPH\" -m 1 --agent flag-agent --prompt test'
	" 2>&1)
	assert_output_contains "$output" "run --agent flag-agent" "Agent flag overrides env/config"
	assert_output_not_contains "$output" "run --agent env-agent" "Env agent is not used when flag exists"
	assert_output_not_contains "$output" "run --agent config-agent" "Config agent is not used when flag exists"
}

# ============================================================================
# Test: Agent is not passed when unset
# ============================================================================
test_agent_not_passed_when_unset() {
	local output
	output=$(run_with_fake_opencode "\"$RALPH\" -m 1 --prompt test")
	assert_output_contains "$output" "run test" "opencode run still executes"
	assert_output_not_contains "$output" "--agent" "--agent is not passed when no agent is configured"
}

# ============================================================================
# Test: Prompt from stdin via --prompt-file -
# ============================================================================
test_prompt_from_stdin_flag() {
	local output
	output=$(printf '%s\n' 'STDIN PROMPT FROM FLAG' | DEBUG=1 "$RALPH" --prompt-file - 2>&1)
	assert_output_contains "$output" "USING PROMPT FROM STDIN" "--prompt-file - reads prompt from stdin"
	assert_output_contains "$output" "STDIN PROMPT FROM FLAG" "stdin prompt content is used with --prompt-file -"
}

# ============================================================================
# Test: Prompt from stdin via positional '-'
# ============================================================================
test_prompt_from_stdin_positional() {
	local output
	output=$(printf '%s\n' 'STDIN PROMPT FROM POSITIONAL' | DEBUG=1 "$RALPH" - 2>&1)
	assert_output_contains "$output" "USING PROMPT FROM STDIN" "positional '-' reads prompt from stdin"
	assert_output_contains "$output" "STDIN PROMPT FROM POSITIONAL" "stdin prompt content is used with positional '-'"
}

# ============================================================================
# Test: Log file flag (short form)
# ============================================================================
test_log_file_short_flag() {
	local tmpdir
	tmpdir=$(mktemp -d)
	trap 'rm -rf $tmpdir' EXIT

	local logfile
	logfile="$tmpdir/ralph.log"

	DEBUG=1 "$RALPH" -l "$logfile" --prompt "log-flag-test" >/dev/null 2>&1

	if [ -f "$logfile" ] && grep -q "log-flag-test" "$logfile"; then
		echo "✓ Log file flag (-l) writes output to log file"
		TESTS_PASSED=$((TESTS_PASSED + 1))
	else
		echo "✗ Log file flag (-l) writes output to log file"
		TESTS_FAILED=$((TESTS_FAILED + 1))
	fi

	rm -rf "$tmpdir"
}

# ============================================================================
# Test: Environment variable - RALPH_MAX_ITERATIONS
# ============================================================================
test_env_max_iterations() {
	local output
	output=$(DEBUG=1 RALPH_MAX_ITERATIONS=10 "$RALPH" 2>&1)
	assert_output_contains "$output" "Max iterations: 10" "RALPH_MAX_ITERATIONS env var sets value"
}

# ============================================================================
# Test: Environment variable - RALPH_SPECS_DIR
# ============================================================================
test_env_specs_dir() {
	local output
	output=$(DEBUG=1 RALPH_SPECS_DIR=my_specs "$RALPH" 2>&1)
	assert_output_contains "$output" "my_specs" "RALPH_SPECS_DIR env var sets value"
}

# ============================================================================
# Test: Environment variable - RALPH_SPECS_INDEX_FILE
# ============================================================================
test_env_specs_index_file() {
	local output
	output=$(DEBUG=1 RALPH_SPECS_INDEX_FILE=OVERVIEW.md "$RALPH" 2>&1)
	assert_output_contains "$output" "OVERVIEW.md" "RALPH_SPECS_INDEX_FILE env var sets value"
}

# ============================================================================
# Test: Environment variable - RALPH_IMPLEMENTATION_PLAN_NAME
# ============================================================================
test_env_impl_plan_name() {
	local output
	output=$(DEBUG=1 RALPH_IMPLEMENTATION_PLAN_NAME=ROADMAP.md "$RALPH" 2>&1)
	assert_output_contains "$output" "ROADMAP.md" "RALPH_IMPLEMENTATION_PLAN_NAME env var sets value"
}

# ============================================================================
# Test: Environment variable - RALPH_LOG_FILE
# ============================================================================
test_env_log_file() {
	local tmpdir
	tmpdir=$(mktemp -d)
	trap 'rm -rf $tmpdir' EXIT

	local logfile
	logfile="$tmpdir/ralph-env.log"

	DEBUG=1 RALPH_LOG_FILE="$logfile" "$RALPH" --prompt "log-env-test" >/dev/null 2>&1

	if [ -f "$logfile" ] && grep -q "log-env-test" "$logfile"; then
		echo "✓ RALPH_LOG_FILE env var writes output to log file"
		TESTS_PASSED=$((TESTS_PASSED + 1))
	else
		echo "✗ RALPH_LOG_FILE env var writes output to log file"
		TESTS_FAILED=$((TESTS_FAILED + 1))
	fi

	rm -rf "$tmpdir"
}

# ============================================================================
# Test: opencode execution output is mirrored to log file
# ============================================================================
test_log_captures_opencode_output() {
	local tmpdir
	tmpdir=$(mktemp -d)
	trap 'rm -rf $tmpdir' EXIT

	local logfile
	logfile="$tmpdir/ralph-opencode.log"

	cat >"$tmpdir/opencode" <<'EOF'
#!/bin/sh
printf '%s\n' 'FAKE-OPENCODE-OUTPUT'
printf '%s\n' '<promise>COMPLETE</promise>'
EOF
	chmod +x "$tmpdir/opencode"

	PATH="$tmpdir:$PATH" "$RALPH" -m 1 -l "$logfile" --prompt "log-opencode-test" >/dev/null 2>&1

	if [ -f "$logfile" ] && grep -q "FAKE-OPENCODE-OUTPUT" "$logfile"; then
		echo "✓ opencode output is written to log file"
		TESTS_PASSED=$((TESTS_PASSED + 1))
	else
		echo "✗ opencode output is written to log file"
		TESTS_FAILED=$((TESTS_FAILED + 1))
	fi

	rm -rf "$tmpdir"
}

# ============================================================================
# Test: Environment variable - RALPH_PROMPTS_DIR
# ============================================================================
test_env_prompts_dir() {
	local tmpdir
	tmpdir=$(mktemp -d)
	trap 'rm -rf $tmpdir' EXIT

	mkdir -p "$tmpdir/custom_prompts"
	cat >"$tmpdir/custom_prompts/build.md" <<EOF
CUSTOM PROMPT FROM ENV VAR DIRECTORY
EOF

	local output
	output=$(cd "$tmpdir" && DEBUG=1 RALPH_PROMPTS_DIR=custom_prompts "$RALPH" build 2>&1)
	assert_output_contains "$output" "USING PROMPT FILE:" "RALPH_PROMPTS_DIR loads prompt file from custom directory"
	assert_output_contains "$output" "CUSTOM PROMPT FROM ENV VAR DIRECTORY" "RALPH_PROMPTS_DIR uses custom prompt content"

	rm -rf "$tmpdir"
}

# ============================================================================
# Test: Config file loading
# ============================================================================
test_config_file_loading() {
	local tmpdir
	tmpdir=$(mktemp -d)
	trap 'rm -rf $tmpdir' EXIT

	# Create a config file
	cat >"$tmpdir/.ralphrc" <<EOF
RALPH_MAX_ITERATIONS=7
RALPH_SPECS_DIR=config_specs
EOF

	local output
	output=$(cd "$tmpdir" && DEBUG=1 "$RALPH" 2>&1)
	assert_output_contains "$output" "Max iterations: 7" "Config file RALPH_MAX_ITERATIONS is loaded"
	assert_output_contains "$output" "config_specs" "Config file RALPH_SPECS_DIR is loaded"

	rm -rf "$tmpdir"
}

# ============================================================================
# Test: Flag overrides environment variable
# ============================================================================
test_flag_overrides_env() {
	local output
	output=$(DEBUG=1 RALPH_MAX_ITERATIONS=20 "$RALPH" -m 5 2>&1)
	assert_output_contains "$output" "Max iterations: 5" "Flag overrides RALPH_MAX_ITERATIONS env var"
}

# ============================================================================
# Test: Flag overrides config file
# ============================================================================
test_flag_overrides_config() {
	local tmpdir
	tmpdir=$(mktemp -d)
	trap 'rm -rf $tmpdir' EXIT

	# Create a config file
	cat >"$tmpdir/.ralphrc" <<EOF
RALPH_MAX_ITERATIONS=15
RALPH_SPECS_DIR=config_specs
EOF

	local output
	output=$(cd "$tmpdir" && DEBUG=1 "$RALPH" -m 8 -s flag_specs 2>&1)
	assert_output_contains "$output" "Max iterations: 8" "Flag overrides config file RALPH_MAX_ITERATIONS"
	assert_output_contains "$output" "flag_specs" "Flag overrides config file RALPH_SPECS_DIR"

	rm -rf "$tmpdir"
}

# ============================================================================
# Test: Environment variable overrides config file
# ============================================================================
test_env_overrides_config() {
	local tmpdir
	tmpdir=$(mktemp -d)
	trap 'rm -rf $tmpdir' EXIT

	# Create a config file
	cat >"$tmpdir/.ralphrc" <<EOF
RALPH_MAX_ITERATIONS=15
RALPH_SPECS_DIR=config_specs
EOF

	local output
	output=$(cd "$tmpdir" && DEBUG=1 RALPH_MAX_ITERATIONS=12 RALPH_SPECS_DIR=env_specs "$RALPH" 2>&1)
	assert_output_contains "$output" "Max iterations: 12" "Env var overrides config file RALPH_MAX_ITERATIONS"
	assert_output_contains "$output" "env_specs" "Env var overrides config file RALPH_SPECS_DIR"

	rm -rf "$tmpdir"
}

# ============================================================================
# Test: Custom config file via flag
# ============================================================================
test_custom_config_file() {
	local tmpdir
	tmpdir=$(mktemp -d)
	trap 'rm -rf $tmpdir' EXIT

	# Create a custom config file
	cat >"$tmpdir/custom.ralphrc" <<EOF
RALPH_MAX_ITERATIONS=9
EOF

	local output
	output=$(cd "$tmpdir" && DEBUG=1 "$RALPH" -c custom.ralphrc 2>&1)
	assert_output_contains "$output" "Max iterations: 9" "Custom config file via -c flag is loaded"

	rm -rf "$tmpdir"
}

# ============================================================================
# Test: Scope placeholder substitution
# ============================================================================
test_scope_placeholder() {
	local output
	output=$(DEBUG=1 "$RALPH" plan "user-auth" 2>&1)
	assert_output_contains "$output" "user-auth" "Scope placeholder is substituted"
}

# ============================================================================
# Test: Unknown flag error
# ============================================================================
test_unknown_flag_error() {
	local output
	output=$("$RALPH" --unknown-flag 2>&1) || true
	assert_output_contains "$output" "Error" "Unknown flag produces error"
}

# ============================================================================
# Test: Missing required flag argument
# ============================================================================
test_missing_flag_argument() {
	local output
	output=$("$RALPH" -m 2>&1) || true
	assert_output_contains "$output" "Error" "Missing required flag argument produces error"
}

# ============================================================================
# Test: DEBUG environment variable
# ============================================================================
test_debug_env_var() {
	local output
	output=$(DEBUG=1 "$RALPH" 2>&1)
	assert_output_contains "$output" "Agent Instructions" "DEBUG=1 prints prompt instead of executing"
}

# ============================================================================
# Run all tests
# ============================================================================
echo ""
echo "Running ralph.sh test suite..."
echo "=============================="
echo ""

test_help_short_flag
test_help_long_flag
test_max_iterations_short_flag
test_max_iterations_long_flag
test_specs_dir_short_flag
test_specs_dir_long_flag
test_specs_index_short_flag
test_specs_index_long_flag
test_no_specs_index_flag
test_impl_plan_short_flag
test_impl_plan_long_flag
test_inline_prompt_flag
test_agent_short_flag
test_agent_long_flag
test_prompt_from_stdin_flag
test_prompt_from_stdin_positional
test_log_file_short_flag
test_env_max_iterations
test_env_specs_dir
test_env_specs_index_file
test_env_impl_plan_name
test_env_agent
test_env_log_file
test_log_captures_opencode_output
test_env_prompts_dir
test_config_file_loading
test_config_agent
test_flag_overrides_env
test_flag_overrides_config
test_env_overrides_config
test_agent_precedence
test_agent_not_passed_when_unset
test_custom_config_file
test_scope_placeholder
test_unknown_flag_error
test_missing_flag_argument
test_debug_env_var

# ============================================================================
# Summary
# ============================================================================
echo ""
echo "=============================="
echo "Test Summary"
echo "=============================="
echo "Passed: $TESTS_PASSED"
if [ $TESTS_FAILED -gt 0 ]; then
	echo "Failed: $TESTS_FAILED"
	exit 1
else
	echo "Failed: 0"
	exit 0
fi
