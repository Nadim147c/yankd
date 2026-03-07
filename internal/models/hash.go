package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"

	"github.com/cespare/xxhash/v2"
)

// Hash is the xxhash sum of clipboard content.
//
// It implements driver.Valuer to map uint64 to SQLite's signed int64, allowing
// the full 64-bit hash to be stored.
type Hash uint64

type hashConstraints interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
	sql.Scanner
	driver.Valuer
	fmt.Stringer
	Int() int64
	Uint() uint64
	Bytes() []byte
}

// do not compile unless all contrains are fulfilled.
var _ hashConstraints = (*Hash)(nil)

// NewHash return xxhash for given data.
func NewHash(b []byte) Hash {
	return Hash(xxhash.Sum64(b))
}

// String implements the [Stringer] interface.
func (h Hash) String() string {
	return strconv.FormatUint(h.Uint(), 16)
}

// Bytes returns BigEndian representation hash bits.
func (h Hash) Bytes() []byte {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], h.Uint())
	return b[:]
}

// Int retruns int64 representation of hash bits.
func (h Hash) Int() int64 {
	return int64(h) //nolint:gosec
}

// Uint retruns uint64 representation of hash bits.
func (h Hash) Uint() uint64 {
	return uint64(h)
}

// Scan implements the [sql.Scanner] interface.
func (h *Hash) Scan(v any) error {
	switch v := v.(type) {
	case int64:
		//nolint:gosec
		*h = Hash(v)
	case uint64:
		*h = Hash(v)
	default:
		return errors.New("invalid error type")
	}
	return nil
}

// Value implements the [driver.Valuer] interface.
func (n Hash) Value() (driver.Value, error) {
	return n.Int(), nil // SQLITE only supports signed 64bit integers!
}

// MarshalText implements [encoding.TextMarshaler] interface.
func (h Hash) MarshalText() (text []byte, err error) {
	return []byte(h.String()), nil
}

// UnmarshalText implements [encoding.TextUnmarshaler] interface.
func (h *Hash) UnmarshalText(text []byte) error {
	v, err := strconv.ParseUint(string(text), 16, 64)
	if err != nil {
		return err
	}
	*h = Hash(v)
	return nil
}
