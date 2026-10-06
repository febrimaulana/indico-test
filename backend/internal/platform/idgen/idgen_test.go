package idgen

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type erringReader struct{}

func (erringReader) Read([]byte) (int, error) { return 0, errors.New("boom") }

func TestNewGeneratesUniqueLookingIDs(t *testing.T) {
	generator := New()
	id, err := generator.NewReservationID()
	require.NoError(t, err)
	require.Regexp(t, `^res_[0-9]{6}$`, id)
}

func TestNewWithReaderDeterministic(t *testing.T) {
	generator := NewWithReader(bytes.NewReader([]byte{0x00, 0x00, 0x00, 0x2A}))
	id, err := generator.NewReservationID()
	require.NoError(t, err)
	require.Equal(t, "res_000042", id)
}

func TestNewWithReaderModulo(t *testing.T) {
	// 0x000F4240 == 1_000_000 -> modulo 1_000_000 == 0.
	generator := NewWithReader(bytes.NewReader([]byte{0x00, 0x0F, 0x42, 0x40}))
	id, err := generator.NewReservationID()
	require.NoError(t, err)
	require.Equal(t, "res_000000", id)
}

func TestNewWithReaderError(t *testing.T) {
	generator := NewWithReader(erringReader{})
	_, err := generator.NewReservationID()
	require.ErrorContains(t, err, "gagal menghasilkan id reservasi")
}
