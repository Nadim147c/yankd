package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type NullString struct {
	String string
	Valid  bool
}

func NewValidString(s string) NullString {
	return NullString{s, true}
}

func NewValidStrings(s []string) []NullString {
	res := make([]NullString, len(s))
	for i := range s {
		res[i] = NewValidString(s[i])
	}
	return res
}

func NewNullString(s string, v bool) NullString {
	return NullString{s, v}
}

func NewNullStringFromPtr(s *string) NullString {
	if s == nil {
		return NullString{}
	}
	return NullString{*s, true}
}

var (
	_ json.Marshaler   = (*NullString)(nil)
	_ json.Unmarshaler = (*NullString)(nil)
	_ sql.Scanner      = (*NullString)(nil)
	_ driver.Valuer    = (*NullString)(nil)
)

// Scan implements the sql.Scanner interface.
func (n *NullString) Scan(v any) error {
	if v == nil {
		n.String = ""
		n.Valid = false
		return nil
	}
	n.Valid = true
	switch v := v.(type) {
	case string:
		n.String = v
	case []byte:
		n.String = string(v)
	default:
		return errors.New("invalid error type")
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (n NullString) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return n.String, nil
}

func (n NullString) MarshalJSON() ([]byte, error) {
	if !n.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(n.String)
}

func (n *NullString) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s == "null" {
		n.String = ""
		n.Valid = false
		return nil
	}
	if err := json.Unmarshal(b, &n.String); err != nil {
		return err
	}
	n.Valid = true
	return nil
}
