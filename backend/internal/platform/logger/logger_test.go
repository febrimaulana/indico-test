package logger

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewLevels(t *testing.T) {
	cases := []struct {
		level        string
		debugEnabled bool
	}{
		{"debug", true},
		{"DEBUG", true},
		{"info", false},
		{"warn", false},
		{"error", false},
		{"", false},
		{"unknown", false},
	}
	for _, tc := range cases {
		t.Run("level "+tc.level, func(t *testing.T) {
			log := New(tc.level)
			require.NotNil(t, log)
			require.Equal(t, tc.debugEnabled, log.Enabled(context.Background(), slog.LevelDebug))
		})
	}
}
