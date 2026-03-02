# Prompts

Status: Proposed

## Overview

### Purpose

- Define how Ralph resolves prompts and generates default build/plan prompts.
- Provide a testable description of prompt precedence and prompt content inputs.

### Goals

- Specify prompt resolution order and failure behavior.
- Document the built-in build and plan prompts at a behavioral level.
- Describe how prompt files are discovered on disk.

### Non-Goals

- Defining new prompt types or templates.
- Editing prompt content beyond what is implemented in code.

### Scope

- In scope: prompt sources, resolution order, default prompt generation.
- Out of scope: agent execution, config precedence (see configuration spec).

## Architecture

### Module/package layout (tree format)

```
internal/
  prompt/
    prompts.go
```

### Component diagram (ASCII)

```
+------------------+
| Prompt Resolver  |
| internal/prompt  |
+---------+--------+
          |
          v
+---------+--------+        +------------------+
| Prompt Content   |<------>| Inline/StdIn     |
+---------+--------+        +------------------+
          |                 +------------------+
          +---------------->| Prompt File(s)   |
          |                 +------------------+
          |                 +------------------+
          +---------------->| Built-in Prompts |
                            +------------------+
```

### Data flow summary

1. Prompt resolver checks for inline prompt text.
2. If not inline, it checks stdin usage.
3. If not stdin, it checks explicit prompt file path.
4. If not explicit, it searches for a prompt file in the prompts directory (walking upward).
5. If not found, it falls back to built-in prompts for `build` and `plan`.
6. If no source is valid, it returns an error.

## Data model

### Core Entities

- PromptSource
  - Enum-like: `Inline`, `Stdin`, `File`, `BuiltIn`.
  - Determines how prompt text is obtained.

- PromptRequest
  - Inputs: `promptName`, `scope`, and config fields `CustomPrompt`, `PromptFile`, `PromptsDir`.
  - Output: prompt text string.

### Relationships

- `PromptRequest` is derived from CLI args and config (see configuration spec for fields).
- `PromptSource` is selected by precedence order.

### Persistence Notes

- Prompt files are plain-text Markdown files on disk.

## Workflows

### Resolve prompt (happy path)

1. If `CustomPrompt` is set, return it.
2. If `PromptFile` is `-` or `promptName` is `-`, read prompt text from stdin.
3. If `PromptFile` is set, read that file.
4. If a prompt file exists at `PromptsDir/<promptName>.md` (searching upward), read it.
5. If `promptName` is `build` or `plan`, generate the built-in prompt.
6. Otherwise, return an error.

### Resolve prompt (missing file)

1. Prompt file path is provided but cannot be read.
2. Resolver returns an error including the path.

### Resolve prompt (unknown name)

1. `promptName` does not match a prompt file or built-in prompt.
2. Resolver returns an error: prompt not found.

## APIs

- None. Prompts are resolved locally.

## Client SDK Design

- Not applicable.

## Configuration

- Prompt resolution uses config fields: `CustomPrompt`, `PromptFile`, `PromptsDir`, `SpecsDir`, `SpecsIndexFile`, `ImplementationPlanName`.
- See configuration spec for full definitions and precedence.

## Permissions

- Requires read access to prompt files and stdin.

## Security Considerations

- Prompt text may include sensitive data; avoid committing secrets to prompt files.
- Prompt content is logged to stdout and optionally to a log file; treat logs as sensitive.

## Dependencies

- Standard library only (`os`, `io`, `path/filepath`).

## Open Questions / Risks

- Should prompt discovery search the current directory before `PromptsDir/<name>.md`?
- Should prompt file lookup be strict to prevent parent directory traversal?

## Verifications

- `ralph --prompt "hello" build` uses inline prompt.
- `echo "hi" | ralph -` reads from stdin.
- `ralph --prompt-file ./prompts/build.md build` uses that file.
- `ralph plan` uses built-in plan prompt when no file exists.

## Appendices

### Built-in prompt behavior (summary)

- Build prompt:
  - Instructs to study specs and the implementation plan.
  - Requires implementing a single task, validating, updating plan, and committing.
  - Includes a completion signal placeholder `<COMPLETION_SIGNAL>`.

- Plan prompt:
  - Instructs to generate/update the implementation plan in a structured format.
  - Requires study/gap analysis against specs and code.
  - Includes a completion signal placeholder `<COMPLETION_SIGNAL>`.
