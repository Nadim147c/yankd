package ipc

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"io"
)

type (
	command int8
	status  int8
)

const (
	commandNull command = iota
	commandEcho
	commandPing
	commandSet
	commandPause
)

const (
	statusOk status = iota
	statusErr
	statusFatal
)

type msg struct {
	command command
	status  status
	buf     []byte
	off     int64
}

type msgContrains interface {
	io.Writer
	io.Reader
	Reset()

	AppendBytes(p []byte)
	AppendString(s string)
	SetBytes(p []byte)
	SetString(s string)
	String() string

	Bytes() []byte
	Header() [2]byte

	Decode(r io.Reader) error
	DecodeBytes([]byte) error
	Encode(r io.Writer) error
	EncodeBytes() []byte
}

var _ msgContrains = (*msg)(nil)

func newErrMsg(e error) *msg {
	m := new(msg)
	m.status = statusErr
	m.SetString(e.Error())
	return m
}

func (m *msg) Write(p []byte) (n int, err error) {
	m.buf = append(m.buf, p...)
	return len(p), nil
}

func (m *msg) Read(b []byte) (n int, err error) {
	if m.off >= int64(len(m.buf)) {
		return 0, io.EOF
	}
	n = copy(b, m.buf[m.off:])
	m.off += int64(n)
	return
}

func (m *msg) Reset() {
	m.off = 0
	m.buf = m.buf[:0]
}

func (m *msg) AppendBytes(p []byte)  { m.buf = append(m.buf, p...) }
func (m *msg) AppendString(s string) { m.buf = append(m.buf, []byte(s)...) }

func (m *msg) SetBytes(p []byte)  { m.buf = p }
func (m *msg) SetString(s string) { m.buf = []byte(s) }

func (m msg) Bytes() []byte  { return m.buf[m.off:] }
func (m msg) String() string { return string(m.Bytes()) }

func (m msg) Header() [2]byte {
	var header [2]byte
	header[0] = byte(m.command)
	header[1] = byte(m.status)
	return header
}

var encoder = base64.StdEncoding

func (m *msg) DecodeBytes(b []byte) error {
	return m.Decode(bytes.NewReader(b))
}

func (m *msg) Decode(r io.Reader) error {
	buf := bufio.NewReader(r)

	cmd, err := buf.ReadByte()
	if err != nil {
		return err
	}
	st, err := buf.ReadByte()
	if err != nil {
		return err
	}

	m.command = command(cmd)
	m.status = status(st)

	p, err := buf.ReadBytes(0)
	m.buf, err = encoder.AppendDecode(nil, dropZeroByte(p))
	return err
}

func dropZeroByte(p []byte) []byte {
	return bytes.TrimRightFunc(p, func(r rune) bool { return r == 0 })
}

func (m msg) EncodeBytes() []byte {
	buf := bytes.NewBuffer(nil)
	m.Encode(buf)
	return buf.Bytes()
}

func (m msg) Encode(w io.Writer) error {
	writer := bufio.NewWriter(w)

	header := m.Header()
	if _, err := writer.Write(header[:]); err != nil {
		return err
	}

	p := encoder.AppendEncode(nil, m.buf)
	if _, err := writer.Write(p); err != nil {
		return err
	}

	writer.WriteByte(0)
	return writer.Flush()
}
