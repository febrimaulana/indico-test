// Package idgen generates collision-resistant reservation identifiers in the
// format res_XXXXXX using crypto/rand.
package idgen

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
)

// Generator produces unique reservation identifiers.
type Generator interface {
	NewReservationID() (string, error)
}

type generator struct {
	reader io.Reader
}

// New returns the production generator backed by crypto/rand.
func New() Generator { return &generator{reader: rand.Reader} }

// NewWithReader allows tests to inject a deterministic reader, including one
// that fails, so error handling can be exercised.
func NewWithReader(reader io.Reader) Generator { return &generator{reader: reader} }

// NewReservationID returns an identifier such as "res_883291".
func (g *generator) NewReservationID() (string, error) {
	buf := make([]byte, 4)
	if _, err := io.ReadFull(g.reader, buf); err != nil {
		return "", fmt.Errorf("gagal menghasilkan id reservasi: %w", err)
	}
	value := binary.BigEndian.Uint32(buf) % 1_000_000
	return fmt.Sprintf("res_%06d", value), nil
}
