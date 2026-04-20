package e2e_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

var coverageMatrixRefPattern = regexp.MustCompile(`TestE2E[\w_]+(?:/[\w_]+)?`)

func collectE2ETestNames() (map[string]struct{}, error) {
	dir, err := e2eDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read e2e dir: %w", err)
	}

	names := make(map[string]struct{})
	for _, entry := range entries {
		if !shouldScanTestFile(entry) {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		if err := collectE2ETestsFromFile(path, names); err != nil {
			return nil, err
		}
	}

	return names, nil
}

func shouldScanTestFile(entry os.DirEntry) bool {
	return !entry.IsDir() && strings.HasSuffix(entry.Name(), "_test.go")
}

func collectE2ETestsFromFile(path string, names map[string]struct{}) error {
	fileSet := token.NewFileSet()
	parsed, err := parser.ParseFile(fileSet, path, nil, parser.SkipObjectResolution)
	if err != nil {
		return fmt.Errorf("parse test file %s: %w", path, err)
	}

	for _, decl := range parsed.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || fn.Name == nil {
			continue
		}

		name := fn.Name.Name
		if strings.HasPrefix(name, "TestE2E") {
			names[name] = struct{}{}
		}
	}

	return nil
}

func collectCoverageMatrixRefs() (map[string]struct{}, error) {
	dir, err := e2eDir()
	if err != nil {
		return nil, err
	}

	matrixPath := filepath.Join(dir, "COVERAGE_MATRIX.md")
	content, err := os.ReadFile(matrixPath)
	if err != nil {
		return nil, fmt.Errorf("read matrix file %s: %w", matrixPath, err)
	}

	matches := coverageMatrixRefPattern.FindAllString(string(content), -1)
	refs := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		base := match
		if idx := strings.Index(match, "/"); idx >= 0 {
			base = match[:idx]
		}
		refs[base] = struct{}{}
	}

	return refs, nil
}

func missingCoverageMappings(e2eTests, matrixRefs map[string]struct{}) []string {
	missing := make([]string, 0)
	for testName := range e2eTests {
		if _, ok := matrixRefs[testName]; !ok {
			missing = append(missing, testName)
		}
	}

	slices.Sort(missing)

	return missing
}

func staleCoverageMappings(e2eTests, matrixRefs map[string]struct{}) []string {
	stale := make([]string, 0)
	for matrixRef := range matrixRefs {
		if _, ok := e2eTests[matrixRef]; !ok {
			stale = append(stale, matrixRef)
		}
	}

	slices.Sort(stale)

	return stale
}

func e2eDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	return wd, nil
}
