package clock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNow(t *testing.T) {
	now := New().Now()
	require.WithinDuration(t, time.Now().UTC(), now, time.Second)
	require.Equal(t, time.UTC, now.Location())
}
