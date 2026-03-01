#!/bin/sh

# Usage: ./ralph.sh [options] [prompt] [scope]
#
# Options:
#    -c, --config FILE                 Config file to source
#    -m, --max-iterations N            Maximum iterations (default: 25)
#    -p, --prompt-file FILE            Prompt file path
#    -s, --specs-dir DIR               Specs directory (default: specs)
#    -i, --specs-index FILE            Specs index file (default: README.md)
#    --no-specs-index                  Disable specs index file
#    -n, --implementation-plan-name N  Implementation plan file name
#    -l, --log-file FILE               Log file path
#    --no-log                          Disable logs
#    --log-truncate                    Truncate log file before writing
#    --stop-condition CONDITION        Custom stop condition text
#    --prompt PROMPT                   Inline custom prompt (overrides prompt files)
#    -h, --help                        Show this help message
#
# Positional Args:
#    prompt: Name of the markdown prompt file or pre-bundled prompts: build|plan. Defaults to "build".
#    scope: Scope of the work to be done (optional)
#
# For extended documentation, examples, and configuration options, visit https://github.com/iyaki/ralph.

set -eu

ralph() ( # Subshell function used to give a scope to code
	# shellcheck disable=SC2034
	OPENCODE_EXPERIMENTAL_PLAN_MODE=0 # Disabled because the opencode experimental plan mode causes hangs on non interactive sessions

	# Function to search for file recursively upwards
	find_file() {
		_file_path="$1"
		_current_dir="$PWD"

		# If it's an absolute path, return as-is
		case "$_file_path" in
		/*)
			echo "$_file_path"
			return
			;;
		esac

		# Search upwards for the file
		while [ "$_current_dir" != "/" ]; do
			if [ -f "$_current_dir/$_file_path" ]; then
				echo "$_current_dir/$_file_path"
				return
			fi
			_current_dir="$(dirname "$_current_dir")"
		done

		# Not found
		echo ""
	}

	# ============================================================================
	# Flag parsing - POSIX compliant
	# ============================================================================

	# Save original environment variables (for precedence: flags > env vars > config > defaults)
	_ORIG_RALPH_MAX_ITERATIONS="${RALPH_MAX_ITERATIONS:-}"
	_ORIG_RALPH_SPECS_DIR="${RALPH_SPECS_DIR:-}"
	_ORIG_RALPH_SPECS_INDEX_FILE="${RALPH_SPECS_INDEX_FILE:-}"
	_ORIG_RALPH_IMPLEMENTATION_PLAN_NAME="${RALPH_IMPLEMENTATION_PLAN_NAME:-}"
	_ORIG_RALPH_CUSTOM_PROMPT="${RALPH_CUSTOM_PROMPT:-}"
	_ORIG_RALPH_LOG_FILE="${RALPH_LOG_FILE:-}"
	_ORIG_RALPH_LOG_ENABLED="${RALPH_LOG_ENABLED:-}"
	_ORIG_RALPH_LOG_APPEND="${RALPH_LOG_APPEND:-}"

	# Initialize flag variables (empty means not set via flag)
	_PROMPT_FILE=""
	_CUSTOM_PROMPT=""
	_CONFIG_FILE=""
	_MAX_ITERATIONS=""
	_SPECS_DIR=""
	_SPECS_INDEX_FILE=""
	_SPECS_INDEX_FILE_DISABLED=""
	_IMPLEMENTATION_PLAN_NAME=""
	_LOG_FILE=""
	_LOG_ENABLED=""
	_LOG_APPEND=""
	PROMPTS_DIR="${RALPH_PROMPTS_DIR:-prompts}"

	# Parse command-line flags
	_args=""
	while [ $# -gt 0 ]; do
		case "$1" in
		-c | --config)
			if [ $# -lt 2 ]; then
				echo "Error: -c/--config requires an argument" >&2
				exit 1
			fi
			_CONFIG_FILE="$2"
			shift 2
			;;
		-m | --max-iterations)
			if [ $# -lt 2 ]; then
				echo "Error: -m/--max-iterations requires an argument" >&2
				exit 1
			fi
			_MAX_ITERATIONS="$2"
			shift 2
			;;
		-p | --prompt-file)
			if [ $# -lt 2 ]; then
				echo "Error: -p/--prompt-file requires an argument" >&2
				exit 1
			fi
			_PROMPT_FILE="$2"
			shift 2
			;;
		-s | --specs-dir)
			if [ $# -lt 2 ]; then
				echo "Error: -s/--specs-dir requires an argument" >&2
				exit 1
			fi
			_SPECS_DIR="$2"
			shift 2
			;;
		-i | --specs-index)
			if [ $# -lt 2 ]; then
				echo "Error: -i/--specs-index requires an argument" >&2
				exit 1
			fi
			_SPECS_INDEX_FILE="$2"
			shift 2
			;;
		--no-specs-index)
			_SPECS_INDEX_FILE_DISABLED="1"
			shift
			;;
		-n | --implementation-plan-name)
			if [ $# -lt 2 ]; then
				echo "Error: -n/--implementation-plan-name requires an argument" >&2
				exit 1
			fi
			_IMPLEMENTATION_PLAN_NAME="$2"
			shift 2
			;;
		-l | --log-file)
			if [ $# -lt 2 ]; then
				echo "Error: -l/--log-file requires an argument" >&2
				exit 1
			fi
			_LOG_FILE="$2"
			shift 2
			;;
		--no-log)
			_LOG_ENABLED="0"
			shift
			;;
		--log-truncate)
			_LOG_APPEND="0"
			shift
			;;
		--prompt)
			if [ $# -lt 2 ]; then
				echo "Error: --prompt requires an argument" >&2
				exit 1
			fi
			_CUSTOM_PROMPT="$2"
			shift 2
			;;
		-h | --help)
			sed -n '3,21p' "$0" | sed 's/^#//' | sed "s%\./ralph\.sh%$0%"
			exit 0
			;;
		--)
			shift
			_args="$_args $*"
			break
			;;
		-*)
			echo "Error: Unknown option: $1" >&2
			exit 1
			;;
		*)
			_args="$_args $1"
			shift
			;;
		esac
	done

	# Find and load config file
	if [ -n "$_CONFIG_FILE" ]; then
		CONFIG_FILE="$_CONFIG_FILE"
	else
		CONFIG_FILE="${RALPH_CONFIG_FILE:-.ralphrc}"
	fi
	CONFIG_FILE="$(find_file "$CONFIG_FILE")"
	if [ -n "$CONFIG_FILE" ]; then
		# shellcheck disable=SC1090
		. "$CONFIG_FILE"
	fi

	# Set final values: flags > env vars > config env vars > defaults
	if [ -n "$_MAX_ITERATIONS" ]; then
		MAX_ITERATIONS="$_MAX_ITERATIONS"
	elif [ -n "$_ORIG_RALPH_MAX_ITERATIONS" ]; then
		MAX_ITERATIONS="$_ORIG_RALPH_MAX_ITERATIONS"
	else
		MAX_ITERATIONS="${RALPH_MAX_ITERATIONS:-25}"
	fi

	if [ -n "$_PROMPT_FILE" ]; then
		PROMPT_FILE="$_PROMPT_FILE"
	else
		PROMPT_FILE="${RALPH_PROMPT_FILE:-}"
	fi

	if [ -n "$_SPECS_DIR" ]; then
		SPECS_DIR="$_SPECS_DIR"
	elif [ -n "$_ORIG_RALPH_SPECS_DIR" ]; then
		SPECS_DIR="$_ORIG_RALPH_SPECS_DIR"
	else
		SPECS_DIR="${RALPH_SPECS_DIR:-specs}"
	fi

	if [ -n "$_SPECS_INDEX_FILE_DISABLED" ]; then
		SPECS_INDEX_FILE=""
	elif [ -n "$_SPECS_INDEX_FILE" ]; then
		SPECS_INDEX_FILE="$_SPECS_INDEX_FILE"
	elif [ -n "$_ORIG_RALPH_SPECS_INDEX_FILE" ]; then
		SPECS_INDEX_FILE="$_ORIG_RALPH_SPECS_INDEX_FILE"
	else
		SPECS_INDEX_FILE="${RALPH_SPECS_INDEX_FILE:-README.md}"
	fi

	if [ -n "$_IMPLEMENTATION_PLAN_NAME" ]; then
		IMPLEMENTATION_PLAN_NAME="$_IMPLEMENTATION_PLAN_NAME"
	elif [ -n "$_ORIG_RALPH_IMPLEMENTATION_PLAN_NAME" ]; then
		IMPLEMENTATION_PLAN_NAME="$_ORIG_RALPH_IMPLEMENTATION_PLAN_NAME"
	else
		IMPLEMENTATION_PLAN_NAME="${RALPH_IMPLEMENTATION_PLAN_NAME:-IMPLEMENTATION_PLAN.md}"
	fi

	if [ -n "$_CUSTOM_PROMPT" ]; then
		CUSTOM_PROMPT="$_CUSTOM_PROMPT"
	elif [ -n "$_ORIG_RALPH_CUSTOM_PROMPT" ]; then
		CUSTOM_PROMPT="$_ORIG_RALPH_CUSTOM_PROMPT"
	else
		CUSTOM_PROMPT="${RALPH_CUSTOM_PROMPT:-}"
	fi

	if [ -n "$_LOG_FILE" ]; then
		LOG_FILE="$_LOG_FILE"
	elif [ -n "$_ORIG_RALPH_LOG_FILE" ]; then
		LOG_FILE="$_ORIG_RALPH_LOG_FILE"
	else
		LOG_FILE="${RALPH_LOG_FILE:-}"
	fi

	if [ -n "$_LOG_ENABLED" ]; then
		LOG_ENABLED="$_LOG_ENABLED"
	elif [ -n "$_ORIG_RALPH_LOG_ENABLED" ]; then
		LOG_ENABLED="$_ORIG_RALPH_LOG_ENABLED"
	else
		LOG_ENABLED="${RALPH_LOG_ENABLED:-1}"
	fi

	if [ -n "$_LOG_APPEND" ]; then
		LOG_APPEND="$_LOG_APPEND"
	elif [ -n "$_ORIG_RALPH_LOG_APPEND" ]; then
		LOG_APPEND="$_ORIG_RALPH_LOG_APPEND"
	else
		LOG_APPEND="${RALPH_LOG_APPEND:-1}"
	fi

	# Configure logging for all ralph output (stdout + stderr)
	if [ "$LOG_ENABLED" = "1" ] && [ -n "$LOG_FILE" ]; then
		LOG_DIR=$(dirname "$LOG_FILE")
		if [ ! -d "$LOG_DIR" ]; then
			mkdir -p "$LOG_DIR"
		fi

		if [ "$LOG_APPEND" != "1" ]; then
			: >"$LOG_FILE"
		fi

		printf '===== Ralph run started at %s =====\n' "$(date '+%Y-%m-%d %H:%M:%S %z')" >>"$LOG_FILE"

		LOG_FIFO="$(mktemp "${TMPDIR:-/tmp}/ralph-log.XXXXXX")"
		rm -f "$LOG_FIFO"
		mkfifo "$LOG_FIFO"

		exec 3>&1 4>&2
		tee -a "$LOG_FILE" <"$LOG_FIFO" &
		LOG_TEE_PID=$!
		exec >"$LOG_FIFO" 2>&1

		cleanup_logging() {
			exec 1>&3 2>&4
			exec 3>&- 4>&-
			wait "$LOG_TEE_PID" 2>/dev/null || true
			rm -f "$LOG_FIFO"
		}
		trap cleanup_logging EXIT
	fi

	build_prompt() {
		SPECS_INDEX_FILE_REFERENCE=""
		if [ "$SPECS_INDEX_FILE" != "" ]; then
			SPECS_INDEX_FILE_REFERENCE="$SPECS_DIR/$SPECS_INDEX_FILE"
		fi

		if [ "$SPECS_INDEX_FILE_REFERENCE" != "" ]; then
			SPECS_INDEX_FILE_REFERENCE=" (including \`$SPECS_INDEX_FILE_REFERENCE\` and related specs)"
		fi

		cat <<EOF
# Agent Instructions (Build Mode)

- Study \`$SPECS_DIR/*\`$SPECS_INDEX_FILE_REFERENCE.
- Study \`$IMPLEMENTATION_PLAN_NAME\` and pick the single most important task.
- Implement the task
- Validate the implementation
- Update the plan
- Commit the changes
- Stop after the commit

## Stop Condition

- After completing the selected task, stop. Do NOT start another task in the same run.
- If ALL stories are complete and passing, reply with:
  \`<COMPLETION_SIGNAL>\`

## IMPORTANT

- Before changes, search the codebase. Do NOT assume functionality is missing.
- Implement ONLY one task. Stop after committing.
- Update \`$IMPLEMENTATION_PLAN_NAME\` when the task is done.
- Use the verification log format: \`YYYY-MM-DD: <command or URL> - <result>\`.
- Keep a \`Manual Deployment Tasks\` section in implementation the plan and use \`None\` when there are no tasks.
- You may implement missing functionality if required, but study relevant \`$SPECS_DIR/*\` first.
- You may add temporary logging as needed and remove if no longer needed.

EOF
	}

	plan_prompt() {
		cat <<EOF
# Agent Instructions (Planning Mode)

Scope: <SCOPE>

## Objective

Generate or update \`$IMPLEMENTATION_PLAN_NAME\` in a structured, phase-based format with:

- Clear status metadata
- Quick reference tables
- Phase sections with paths and checklists
- Verification log entries
- Summary tables and remaining effort

Plan only. Do NOT implement anything.

## Study and Gap Analysis

- Study \`$SPECS_DIR/*\` to learn application requirements.
- Study \`$IMPLEMENTATION_PLAN_NAME\` (if present; it may be incorrect).
- Study relevant source code to compare against specs.
- Use \`git\` to study recent changes on the specs related to the specified current scope.

Rules:

- Do NOT assume missing; confirm via code search first.
- Identify where work already exists, partial implementations, TODOs, placeholders, skipped/flaky tests, or inconsistent patterns.
- Keep the plan concise but complete; prefer lists and tables over paragraphs.
- Use \`[x]\` only when verified in code. Use \`[ ]\` if missing or unverified.
- Regenerate the plan if it becomes stale, contradictory, or significantly out of sync with code.
- If the specified scope has relationships with other domain areas, implementation may be needed in those areas as well (always study the related specs and code). Include this in the plan.

## Output Format Requirements

Write \`$IMPLEMENTATION_PLAN_NAME\` using this structure and level of detail:

Header

- Title: \`Implementation Plan (<Scope>)\`
- Status line: \`**Status:** <summary (e.g., "UI Components Complete (39/39)")>\`
- Last Updated date: \`YYYY-MM-DD\`
- Reference to primary spec(s)

Quick Reference

- A table mapping systems/subsystems to:
  - Specs
  - Modules/packages
  - Web packages
  - Migrations or other artifacts
- Use \`✅\` to mark items already implemented.

Phased Plan

- Use numbered phases (e.g., Phase 9, Phase 10) aligned to the spec's domain.
- Each phase includes:
  - Goal
  - Status (if applicable)
  - Paths (directories or file patterns)
  - Checklist with \`[x]\` for verified complete and \`[ ]\` for missing
  - Definition of Done (tests run, commands/URLs, files touched)
  - Risks/Dependencies (brief)
- Break phases into subsections (e.g., 9.1, 9.2) with scope-specific paths and item lists.
- Include "Reference pattern" links when there's a canonical directory or file to follow.

Verification Log

- A chronological log of verification steps with dates.
- Each entry includes:
  - What was verified (endpoints, commands, builds, tests, UI routes)
  - Exact commands or URLs used
  - Tests run and results
  - Bug fixes discovered (if any)
  - Files touched (if known from code search)
  - Use format: \`YYYY-MM-DD: <command or URL> - <result>\`

Summary

- Table of phases with completion status
  - "Remaining effort" line summarizing unfinished sections

Known Existing Work

- Brief section listing confirmed existing implementations to prevent duplicate work

Manual Deployment Tasks

- Required section to document manual steps needed before or during production deployment (manual configuration, third-party service setup, API key acquisition, etc).
- If not applicable, write exactly: \`None\`.

## Stop Condition

**IMPORTANT**: After writing/updating if \`$IMPLEMENTATION_PLAN_NAME\` already reflects the current gaps, reply with:
\`<COMPLETION_SIGNAL>\`

EOF
	}

	PROMPT=""

	# Parse positional arguments
	_PROMPT_NAME="build"
	_SCOPE="Whole system"
	_idx=1
	for _arg in $_args; do
		if [ $_idx -eq 1 ]; then
			_PROMPT_NAME="$_arg"
		elif [ $_idx -eq 2 ]; then
			_SCOPE="$_arg"
		fi
		_idx=$((_idx + 1))
	done

	# If custom prompt is provided via flag, use it directly
	if [ -n "$CUSTOM_PROMPT" ]; then
		PROMPT="$CUSTOM_PROMPT"
		echo ""
		echo "==============================================================="
		echo "               USING INLINE CUSTOM PROMPT"
		echo "==============================================================="
		echo ""
	else


		if [ -n "$_PROMPT_FILE" ]; then
			PROMPT_FILE_PROVIDED="$_PROMPT_FILE"
		else
			# Build prompt file path
			PROMPT_FILE_PROVIDED="${PROMPTS_DIR}/${_PROMPT_NAME}.md"
			# Resolve prompt file location
			PROMPT_FILE="$(find_file "$PROMPT_FILE_PROVIDED")"
		fi

		if [ "$PROMPT_FILE" = "" ]; then
			case "$_PROMPT_NAME" in
			build)
				PROMPT="$(build_prompt)"
				echo ""
				echo "==============================================================="
				echo "               USING DEFAULT 'BUILD' PROMPT"
				echo "==============================================================="
				echo ""
				;;
			plan)
				PROMPT="$(plan_prompt)"
				echo ""
				echo "==============================================================="
				echo "               USING DEFAULT 'PLAN' PROMPT"
				echo "==============================================================="
				echo ""
				;;
			*)
				echo ""
				echo "==============================================================="
				printf " Error: Prompt file not found for '%s'\n Use a valid prompt file or one of the pre-bundled prompts (build, plan).\n" "$PROMPT_FILE_PROVIDED"
				echo "==============================================================="
				echo ""
				exit 1
				;;
			esac
		else
			PROMPT="$(cat "$PROMPT_FILE")"
			echo ""
			echo "==============================================================="
			echo " USING PROMPT FILE: $PROMPT_FILE"
			echo "==============================================================="
			echo ""
		fi
	fi

	COMPLETION_SIGNAL="<promise>COMPLETE</promise>"

	PROMPT="$(echo "$PROMPT" | sed "s|<COMPLETION_SIGNAL>|$COMPLETION_SIGNAL|g")"
	PROMPT="$(echo "$PROMPT" | sed "s|<SCOPE>|$_SCOPE|g")"

	echo "Starting Ralph - Max iterations: $MAX_ITERATIONS"

	i=1
	while [ "$i" -le "$MAX_ITERATIONS" ]; do
		echo ""
		echo "==============================================================="
		echo " [${_PROMPT_NAME}] Iteration $i of $MAX_ITERATIONS ($(date))"
		echo "==============================================================="

		if [ -z "$DEBUG" ]; then
			# For usage with any other agent, just change this line to call the appropriate command with the prompt as input
			OUTPUT="$(opencode run "$PROMPT" 2>&1)" || true
		else
			echo "$PROMPT"
			OUTPUT="$COMPLETION_SIGNAL"
		fi

		# Check for completion signal
		if echo "$OUTPUT" | grep -q "$COMPLETION_SIGNAL"; then

			echo ""
			echo "All planned tasks completed!"
			echo "Completed at iteration $i of $MAX_ITERATIONS"

			break
		fi

		echo "Iteration $i complete. Continuing..."
		i=$((i + 1))
	done

	if [ "$i" -gt "$MAX_ITERATIONS" ]; then
		echo ""
		echo "Reached max iterations ($MAX_ITERATIONS) without completing all planned tasks."
		exit 1
	fi
)

if ralph "$@"; then
	unset -f ralph
else
	unset -f ralph
	exit 1
fi
