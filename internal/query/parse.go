package query

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/dromara/carbon/v2"
)

type Flag uint8

const (
	Fuzzy  Flag = 0b_00
	Suffix Flag = 0b_01
	Prefix Flag = 0b_10
)

type Query struct {
	// Before indicates item after specified time
	After *carbon.Carbon
	// Before indicates item before specified time
	Before *carbon.Carbon
	// Type is the mime-type
	Type string
	// Fuzzy is the primary query
	Fuzzy string
	// Regex is pattern to match
	Regex string
	// Keywords to exact matches
	Keywords []string
	// Flag is search flags (prefix/suffix)
	Flag Flag
	// Limit is limit of search result
	Limit int64
}

func parseHumanTime(s string) *carbon.Carbon {
	t := carbon.Parse(s)
	if !t.HasError() {
		return t
	}
	out, err := exec.Command("date", "-d", s, "+%s").Output() //nolint:noctx
	if err != nil {
		return t
	}
	var ts int64
	_, err = fmt.Sscanf(string(out), "%d", &ts)
	if err != nil {
		return t
	}
	return carbon.CreateFromTimestamp(ts)
}

// Parse always returns a query, any propagated errors are treated as warnings.
func Parse(query []byte) (res Query, warnings []error) {
	lastIndex := len(query) - 1

	fuzzyQuery := bytes.NewBuffer(nil)
	buf := bytes.NewBuffer(nil)
	for i := 0; i < len(query); i++ {
		char := query[i]

		if isQouteOrRegex(char) && lastIndex != i {
			buf.Reset()
			n, ended := collect(buf, query[i+1:], char)
			if ended {
				switch char {
				case '\'', '"', '`':
					res.Keywords = append(res.Keywords, buf.String())
				case '/':
					res.Regex = buf.String()
				default:
					panic("unreachable")
				}
			} else {
				fuzzyQuery.Write(query[i:])
			}
			i += n
			continue
		}

		if bytes.HasPrefix(query[i:], []byte("after:")) { //nolint
			newIdx := i + len("after:")
			if newIdx >= len(query) {
				break
			}

			var dateString string

			if qc := query[newIdx]; isQuote(qc) {
				newIdx++
				if newIdx >= len(query) {
					break
				}
				buf.Reset()
				n, _ := collect(buf, query[newIdx:], qc)
				dateString = buf.String()
				i = newIdx + n - 1
			} else {
				idx := bytes.IndexByte(query[newIdx:], ' ')
				if idx >= 0 {
					dateString = string(query[newIdx : newIdx+idx])
					i = newIdx + idx
				} else {
					dateString = string(query[newIdx:])
					i = lastIndex
				}
			}

			d := parseHumanTime(dateString)
			if d.HasError() {
				warnings = append(warnings, d.Error)
			}
			res.After = d
			continue
		}

		if bytes.HasPrefix(query[i:], []byte("before:")) { //nolint
			newIdx := i + len("before:")
			if newIdx >= len(query) {
				break
			}

			var date string

			if qc := query[newIdx]; isQuote(qc) {
				newIdx++
				if newIdx >= len(query) {
					break
				}
				buf.Reset()
				n, _ := collect(buf, query[newIdx:], qc)
				date = buf.String()
				i = newIdx + n - 1
			} else {
				idx := bytes.IndexByte(query[newIdx:], ' ')
				if idx >= 0 {
					date = string(query[newIdx : newIdx+idx])
					i = newIdx + idx
				} else {
					date = string(query[newIdx:])
					i = lastIndex
				}
			}

			d := parseHumanTime(date)
			if d.HasError() {
				warnings = append(warnings, d.Error)
			}
			res.Before = d
			continue
		}

		if bytes.HasPrefix(query[i:], []byte("type:")) { //nolint
			newIdx := i + len("type:")
			if newIdx >= len(query) {
				break
			}

			var typ string

			if qc := query[newIdx]; isQuote(qc) {
				newIdx++
				if newIdx >= len(query) {
					break
				}
				buf.Reset()
				n, _ := collect(buf, query[newIdx:], qc)
				typ = buf.String()
				i = newIdx + n - 1
			} else {
				idx := bytes.IndexByte(query[newIdx:], ' ')
				if idx >= 0 {
					typ = string(query[newIdx : newIdx+idx])
					i = newIdx + idx
				} else {
					typ = string(query[newIdx:])
					i = lastIndex
				}
			}

			res.Type = typ

			continue
		}

		fuzzyQuery.WriteByte(char)
	}

	b := fuzzyQuery.String()
	flag := Fuzzy
	if after, ok := strings.CutPrefix(b, "["); ok {
		b = after
		flag |= Prefix
	}
	if before, ok := strings.CutSuffix(b, "]"); ok {
		b = before
		flag |= Suffix
	}

	if flag == Prefix|Suffix {
		flag = Fuzzy
	}

	if flag == Fuzzy {
		b = strings.TrimSpace(b)
	}

	res.Fuzzy = b
	res.Flag = flag
	return res, warnings
}

func isQouteOrRegex(char byte) bool {
	return isQuote(char) || isRegex(char)
}

func isQuote(char byte) bool {
	return char == '\'' || char == '"' || char == '`'
}

func isRegex(char byte) bool {
	return char == '/'
}

func collect(buf *bytes.Buffer, b []byte, toEscape byte) (n int, ended bool) {
	for i := 0; i < len(b); i++ {
		char := b[i]
		if char == toEscape {
			return i + 1, true
		}
		if char == '\\' && i+1 < len(b) && b[i+1] == toEscape {
			buf.WriteByte(toEscape)
			i++
			continue
		}
		buf.WriteByte(char)
	}
	return len(b), false
}
