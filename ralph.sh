#!/bin/sh

# Usage: ./ralph.sh [prompt] [scope]
# Params:
#    prompt: Name of the markdown prompt file or one of the pre-bundled prompts: build|plan. Defaults to "build".
#    scope: Scope of the work to be done (optional)

set -eu

ralph() ( # Subshell function used to give a scope to code
	CONFIG_FILE=""
	if [ -f "$CONFIG_FILE" ]; then
		# shellcheck disable=SC1090
		. "$CONFIG_FILE"
	fi

	MAX_ITERATIONS=${RALPH_MAX_ITERATIONS:-25}
	PROMPTS_DIR=${RALPH_PROMPTS_DIR:-prompts}
	# shellcheck disable=SC2034
	OPENCODE_EXPERIMENTAL_PLAN_MODE=0 # Disabled because the opencode experimental plan mode causes hangs on non interactive sessions
	SPECS_DIR=${RALPH_SPECS_DIR:-specs}
	SPECS_INDEX_FILE=${RALPH_SPECS_INDEX_FILE:-README.md}
	IMPLEMENTATION_PLAN_FILE=${RALPH_IMPLEMENTATION_PLAN_FILE:-IMPLEMENTATION_PLAN.md}
	STOP_CONDITION="${RALPH_STOP_CONDITION:-}"
	STOP_CONDITION_SIGNAL="<promise>COMPLETE</promise>"

	# Function to search for prompt file recursively upwards
	find_prompt_file() {
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
- Study \`$IMPLEMENTATION_PLAN_FILE\` and pick the single most important task.
- Implement the task
- Validate the implementation
- Update the plan
- Commit the changes
- Stop after the commit

## Stop Condition

- After completing the selected task, stop. Do NOT start another task in the same run.
- If ALL stories are complete and passing, reply with:
  \`<STOP_CONDITION_SIGNAL>\`

## IMPORTANT

- Before changes, search the codebase. Do NOT assume functionality is missing.
- Implement ONLY one task. Stop after committing.
- Update \`$IMPLEMENTATION_PLAN_FILE\` when the task is done.
- Use the verification log format: \`YYYY-MM-DD: <command or URL> - <result>\`.
- Keep a \`Manual Deployment Tasks\` section in implementation the plan and use \`None\` when there are no tasks.
- You may implement missing functionality if required, but study relevant \`$SPECS_DIR/*\` first.
- You may add temporary logging as needed and remove if no longer needed.

EOF
	}

	plan_prompt() {
		cat <<EOF
# Agent Instructions (Planning Mode)

## Objective

Generate or update \`$IMPLEMENTATION_PLAN_FILE\` in a structured, phase-based format with:

- Clear status metadata
- Quick reference tables
- Phase sections with paths and checklists
- Verification log entries
- Summary tables and remaining effort

Plan only. Do NOT implement anything.

## Study and Gap Analysis

- Study \`$SPECS_DIR/*\` to learn application requirements.
- Study \`$IMPLEMENTATION_PLAN_FILE\` (if present; it may be incorrect).
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

Write \`$IMPLEMENTATION_PLAN_FILE\` using this structure and level of detail:

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

**IMPORTANT**: After writing/updating if \`$IMPLEMENTATION_PLAN_FILE\` already reflects the current gaps, reply with:
\`<STOP_CONDITION_SIGNAL>\`

EOF
	}

	PROMPT=""

	# Parse arguments
	PROMPT_NAME=${1:-build}

	# Build prompt file path
	PROMPT_FILE_PROVIDED="${PROMPTS_DIR}/${PROMPT_NAME}.md"

	# Resolve prompt file location
	PROMPT_FILE="$(find_prompt_file "$PROMPT_FILE_PROVIDED")"

	echo ""
	echo "==============================================================="
	if [ "$PROMPT_FILE" = "" ]; then
		case "$PROMPT_NAME" in
		build)
			echo "               USING DEFAULT 'BUILD' PROMPT"
			PROMPT="$(build_prompt)"
			;;
		plan)
			echo "               USING DEFAULT 'PLAN' PROMPT"
			PROMPT="$(plan_prompt)"
			;;
		*)
			printf " Error: Prompt file not found for '%s'\n Use a valid prompt file or one of the pre-bundled prompts (build, plan).\n" "$PROMPT_FILE_PROVIDED"
			echo "==============================================================="
			echo ""
			exit 1
			;;
		esac
	else
		echo " USING PROMPT FILE: $PROMPT_FILE"
		PROMPT="$(cat "$PROMPT_FILE")"
	fi
	echo "==============================================================="
	echo ""

	if [ -n "$STOP_CONDITION" ]; then
		PROMPT="$PROMPT
**IMPORTANT Stop Condition**: $STOP_CONDITION."
	fi

	PROMPT="$(echo "$PROMPT" | sed "s|<STOP_CONDITION_SIGNAL>|$STOP_CONDITION_SIGNAL|g")"

	SCOPE="${2:-}"
	if [ -n "$SCOPE" ]; then
		PROMPT="${PROMPT}
**Current Scope**: <CURRENT_SCOPE>."
	fi

	PROMPT="$(echo "$PROMPT" | sed "s|<CURRENT_SCOPE>|$SCOPE|g")"

	echo "Starting Ralph - Max iterations: $MAX_ITERATIONS"

	i=1
	while [ "$i" -le "$MAX_ITERATIONS" ]; do
		echo ""
		echo "==============================================================="
		echo " [${PROMPT_NAME}] Iteration $i of $MAX_ITERATIONS ($(date))"
		echo "==============================================================="

		# OUTPUT="$(opencode run "$PROMPT" 2>&1 | tee /dev/stderr)" || true
		echo "$PROMPT"
		OUTPUT="$STOP_CONDITION_SIGNAL"

		# Check for completion signal
		if echo "$OUTPUT" | grep -q "$STOP_CONDITION_SIGNAL"; then

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
