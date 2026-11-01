package models

import (
	"errors"
	"io"
)

var ErrTooBig = errors.New("data exceeds allowed limit")

type StrictLimitReader struct {
	r     io.Reader
	n     int64
	limit int64
}

func NewStrictLimitReader(r io.Reader, limit int64) io.Reader {
	return &StrictLimitReader{
		r:     r,
		n:     limit,
		limit: limit,
	}
}

func (l *StrictLimitReader) Read(p []byte) (int, error) {
	if l.n < 0 {
		return 0, ErrTooBig
	}
	if int64(len(p)) > l.n+1 {
		p = p[0 : l.n+1]
	}
	n, err := l.r.Read(p)
	l.n -= int64(n)
	if l.n < 0 {
		return n, ErrTooBig
	}
	return n, err
}
