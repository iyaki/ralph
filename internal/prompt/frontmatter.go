package prompt

import (
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	frontMatterDelim    = "---"
	frontMatterDelimLen = len(frontMatterDelim)
)

// FrontMatterSettings holds the configuration extracted from the front matter.
type FrontMatterSettings struct {
	Model     string `yaml:"model,omitempty"`
	AgentMode string `yaml:"agent-mode,omitempty"`
}

// ParseFrontMatter parses the YAML front matter from a markdown string.
// It returns the parsed settings, the body with front matter stripped, and an error if parsing fails.
func ParseFrontMatter(content string) (*FrontMatterSettings, string, error) {
	// Front matter must start with "---" at the very beginning of the file.
	if !strings.HasPrefix(content, frontMatterDelim) {
		return &FrontMatterSettings{}, content, nil
	}

	// It must be followed by a newline (or end of file, though empty front matter is weird)
	// If content is just "---", it's not valid front matter with a body usually, but let's check length.
	if len(content) > frontMatterDelimLen {
		charAfter := content[frontMatterDelimLen]
		if charAfter != '\n' && charAfter != '\r' {
			// "---" is followed by something other than newline, e.g. "---foo"
			return &FrontMatterSettings{}, content, nil
		}
	} else {
		// Content is exactly "---", no closing, treat as text
		return &FrontMatterSettings{}, content, nil
	}

	// Search for the closing delimiter.
	// We look for "\n---". This covers both "\n---" and "\r\n---" (partially).
	rest := content[frontMatterDelimLen:]
	idxLF := strings.Index(rest, "\n---")

	if idxLF == -1 {
		// No closing delimiter found
		return &FrontMatterSettings{}, content, nil
	}

	closeIdx := idxLF
	delimLen := 4

	// Check if it is CRLF
	if idxLF > 0 && rest[idxLF-1] == '\r' {
		closeIdx = idxLF - 1
		delimLen = 5
	}

	// Extract front matter content
	frontMatter := rest[:closeIdx]

	// Extract body
	// The body starts after "\n---" or "\r\n---"
	bodyStart := frontMatterDelimLen + closeIdx + delimLen
	body := ""
	if bodyStart < len(content) {
		body = content[bodyStart:]
	}

	var settings FrontMatterSettings
	if err := yaml.Unmarshal([]byte(frontMatter), &settings); err != nil {
		return nil, "", err
	}

	return &settings, strings.TrimSpace(body), nil
}
