package e2e_test

import "testing"

func TestCoverageMatrixCompleteness(t *testing.T) {
	e2eTests, err := collectE2ETestNames()
	if err != nil {
		t.Fatalf("failed to collect e2e tests: %v", err)
	}

	matrixRefs, err := collectCoverageMatrixRefs()
	if err != nil {
		t.Fatalf("failed to collect coverage matrix refs: %v", err)
	}

	missing := missingCoverageMappings(e2eTests, matrixRefs)
	if len(missing) > 0 {
		t.Fatalf("missing coverage matrix mappings: %v", missing)
	}

	stale := staleCoverageMappings(e2eTests, matrixRefs)
	if len(stale) > 0 {
		t.Fatalf("stale coverage matrix mappings: %v", stale)
	}
}
