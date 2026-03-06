package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"errors"
	"math"
	"strconv"

	"github.com/cespare/xxhash/v2"
)

type Hash int64

var (
	_ encoding.TextMarshaler   = (*Hash)(nil)
	_ encoding.TextUnmarshaler = (*Hash)(nil)
	_ sql.Scanner              = (*Hash)(nil)
	_ driver.Valuer            = (*Hash)(nil)
)

// NewHash return xxhash for given data.
func NewHash(b []byte) Hash {
	u := xxhash.Sum64(b)
	return Hash(u % math.MaxInt64) //nolint:gosec
}

// String implements the [Stringer] interface.
func (h Hash) String() string {
	return strconv.FormatInt(h.Int(), 16)
}

// String implements the [Stringer] interface.
func (h Hash) Int() int64 {
	return int64(h)
}

// Scan implements the sql.Scanner interface.
func (h *Hash) Scan(v any) error {
	switch v := v.(type) {
	case int64:
		*h = Hash(v)
	case uint64:
		*h = Hash(v) //nolint:gosec
	default:
		return errors.New("invalid error type")
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (n Hash) Value() (driver.Value, error) {
	return n.Int(), nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (h Hash) MarshalText() (text []byte, err error) {
	return []byte(h.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (h *Hash) UnmarshalText(text []byte) error {
	v, err := strconv.ParseInt(string(text), 16, 64)
	if err != nil {
		return err
	}
	*h = Hash(v)
	return nil
}
