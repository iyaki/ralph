package cli

import (
	"testing"

	"github.com/iyaki/ralph/internal/config"
)

func TestReadBoolFlagOverrideReturnsZeroValueWhenUnchanged(t *testing.T) {
	cmd := NewRunCommand()

	override, err := readBoolFlagOverride(cmd, "no-log")
	if err != nil {
		t.Fatalf("expected no error for unchanged flag, got %v", err)
	}
	if override.changed {
		t.Fatalf("expected unchanged override, got %+v", override)
	}
}

func TestReadBoolFlagOverrideTracksExplicitFalse(t *testing.T) {
	cmd := NewRunCommand()
	if err := cmd.ParseFlags([]string{"--no-log=false"}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	override, err := readBoolFlagOverride(cmd, "no-log")
	if err != nil {
		t.Fatalf("expected no error reading override, got %v", err)
	}
	if !override.changed {
		t.Fatalf("expected changed override, got %+v", override)
	}
	if override.value {
		t.Fatalf("expected override value false, got %+v", override)
	}
}

func TestApplyBoolFlagOverridesAppliesOnlyChangedFlags(t *testing.T) {
	tests := []struct {
		name                string
		initialNoLog        bool
		initialLogTruncate  bool
		noLogOverride       boolFlagOverride
		logTruncateOverride boolFlagOverride
		expectedNoLog       bool
		expectedLogTruncate bool
	}{
		{
			name:               "both changed",
			initialNoLog:       true,
			initialLogTruncate: false,
			noLogOverride: boolFlagOverride{
				changed: true,
				value:   false,
			},
			logTruncateOverride: boolFlagOverride{
				changed: true,
				value:   true,
			},
			expectedNoLog:       false,
			expectedLogTruncate: true,
		},
		{
			name:                "unchanged overrides",
			initialNoLog:        true,
			initialLogTruncate:  false,
			expectedNoLog:       true,
			expectedLogTruncate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				NoLog:       tt.initialNoLog,
				LogTruncate: tt.initialLogTruncate,
			}

			applyBoolFlagOverrides(cfg, tt.noLogOverride, tt.logTruncateOverride)

			if cfg.NoLog != tt.expectedNoLog {
				t.Fatalf("expected NoLog=%v, got %v", tt.expectedNoLog, cfg.NoLog)
			}
			if cfg.LogTruncate != tt.expectedLogTruncate {
				t.Fatalf("expected LogTruncate=%v, got %v", tt.expectedLogTruncate, cfg.LogTruncate)
			}
		})
	}
}
