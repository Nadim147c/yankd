package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/zeebo/xxh3"
)

// Hash is the xxhash sum of clipboard content.
//
// Hash scans and values as DuckDB's UHUGEINT (unsigned 128 bit integer).
type Hash [16]byte

type hashConstraints interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
	sql.Scanner
	driver.Valuer
	fmt.Stringer
	Uint128() xxh3.Uint128
	BigInt() *big.Int
	Bytes() []byte
}

// Do not compile unless all constrains are fulfilled.
var _ hashConstraints = (*Hash)(nil)

// NewHash return xxhash for given data.
func NewHash(b []byte) Hash {
	return Hash(xxh3.Hash128(b).Bytes())
}

// String implements the [Stringer] interface.
func (h Hash) String() string {
	return base64.RawStdEncoding.EncodeToString(h[:])
}

// Bytes returns BigEndian representation hash bits.
func (h Hash) Bytes() []byte {
	return h[:]
}

// Bytes returns BigEndian representation hash bits.
func (h Hash) Uint128() (v xxh3.Uint128) {
	v.Hi = binary.BigEndian.Uint64(h[:8])
	v.Lo = binary.BigEndian.Uint64(h[8:16])
	return v
}

// Bytes returns BigEndian representation hash bits.
func (h Hash) BigInt() *big.Int {
	hi := binary.BigEndian.Uint64(h[:8])
	lo := binary.BigEndian.Uint64(h[8:16])

	bigHi := new(big.Int).SetUint64(hi)
	bigHi.Lsh(bigHi, 64)

	bigLo := new(big.Int).SetUint64(lo)
	return new(big.Int).Or(bigHi, bigLo)
}

// Scan implements the [sql.Scanner] interface.
func (h *Hash) Scan(v any) error {
	switch v := v.(type) {
	case *big.Int:
		*h = Hash(pad128(v.Bytes()))
	case []byte:
		*h = Hash(pad128(v))
	default:
		return errors.New("invalid error type")
	}
	return nil
}

// pad128 left-pads or truncates b to exactly 16 bytes.
func pad128(b []byte) []byte {
	if len(b) == 16 {
		return b
	}
	p := make([]byte, 16)
	if len(b) > 16 {
		copy(p[:], b[len(b)-16:])
	} else {
		copy(p[16-len(b):], b)
	}
	return p
}

// Value implements the [driver.Valuer] interface.
func (h Hash) Value() (driver.Value, error) {
	return h.BigInt(), nil
}

// MarshalText implements [encoding.TextMarshaler] interface.
func (h Hash) MarshalText() (text []byte, err error) {
	return json.Marshal(h.String())
}

// UnmarshalText implements [encoding.TextUnmarshaler] interface.
func (h *Hash) UnmarshalText(text []byte) error {
	var s string
	err := json.Unmarshal(text, &s)
	if err != nil {
		return err
	}
	b, err := base64.RawStdEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	*h = Hash(b)
	return nil
}
