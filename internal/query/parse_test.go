package query

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/dromara/carbon/v2"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		want     Query
		wantErrs int // number of expected warnings
	}{
		{
			name:  "empty query",
			input: "",
			want: Query{
				Fuzzy: "",
				Flag:  Fuzzy,
			},
		},
		{
			name:  "simple fuzzy search",
			input: "hello world",
			want: Query{
				Fuzzy: "hello world",
				Flag:  Fuzzy,
			},
		},
		{
			name:  "prefix fuzzy search",
			input: "[hello",
			want: Query{
				Fuzzy: "hello",
				Flag:  Prefix,
			},
		},
		{
			name:  "suffix fuzzy search",
			input: "hello]",
			want: Query{
				Fuzzy: "hello",
				Flag:  Suffix,
			},
		},
		{
			name:  "exact match (prefix and suffix brackets)",
			input: "[hello]",
			want: Query{
				Fuzzy: "hello",
				Flag:  Fuzzy, // both brackets cancel out
			},
		},
		{
			name:  "single-quoted keyword",
			input: "'hello world'",
			want: Query{
				Fuzzy:    "",
				Keywords: []string{"hello world"},
				Flag:     Fuzzy,
			},
		},
		{
			name:  "double-quoted keyword",
			input: `"hello world"`,
			want: Query{
				Fuzzy:    "",
				Keywords: []string{"hello world"},
				Flag:     Fuzzy,
			},
		},
		{
			name:  "backtick-quoted keyword",
			input: "`hello world`",
			want: Query{
				Fuzzy:    "",
				Keywords: []string{"hello world"},
				Flag:     Fuzzy,
			},
		},
		{
			name:  "regex query",
			input: "/hello.*world/",
			want: Query{
				Fuzzy: "",
				Regex: "hello.*world",
				Flag:  Fuzzy,
			},
		},
		{
			name:  "multiple keywords with fuzzy text",
			input: "fuzzy 'exact one' more `exact two` end",
			want: Query{
				Fuzzy:    "fuzzy  more  end",
				Keywords: []string{"exact one", "exact two"},
				Flag:     Fuzzy,
			},
		},
		{
			name:  "after with unquoted date",
			input: "after:2024-01-15",
			want: Query{
				Fuzzy: "",
				Flag:  Fuzzy,
				After: carbon.CreateFromDate(2024, 1, 15),
			},
		},
		{
			name:  "before with unquoted date",
			input: "before:2024-12-31",
			want: Query{
				Fuzzy:  "",
				Flag:   Fuzzy,
				Before: carbon.CreateFromDate(2024, 12, 31),
			},
		},
		{
			name:  "after with quoted date",
			input: "after:'2024-01-15'",
			want: Query{
				Fuzzy: "",
				Flag:  Fuzzy,
				After: carbon.CreateFromDate(2024, 0o1, 15),
			},
		},
		{
			name:  "before with quoted date",
			input: `before:"2024-12-31"`,
			want: Query{
				Fuzzy:  "",
				Flag:   Fuzzy,
				Before: carbon.CreateFromDate(2024, 12, 31),
			},
		},
		{
			name:  "after and before with fuzzy",
			input: "after:'2024-01-01' before:'2024-12-31' christmas",
			want: Query{
				Fuzzy:  "christmas",
				Flag:   Fuzzy,
				After:  carbon.CreateFromDate(2024, 1, 1),
				Before: carbon.CreateFromDate(2024, 12, 31),
			},
		},
		{
			name:  "escaped quote inside keyword",
			input: `'it\'s working'`,
			want: Query{
				Fuzzy:    "",
				Keywords: []string{"it's working"},
				Flag:     Fuzzy,
			},
		},
		{
			name:  "escaped backtick inside backtick keyword",
			input: "`back\\`tick`",
			want: Query{
				Fuzzy:    "",
				Keywords: []string{"back`tick"},
				Flag:     Fuzzy,
			},
		},
		{
			name:  "escaped slash inside regex",
			input: `/path\/to\/file/`,
			want: Query{
				Fuzzy: "",
				Regex: "path/to/file",
				Flag:  Fuzzy,
			},
		},
		{
			name:  "fuzzy with mixed content",
			input: "prefix[ 'exact' after:'2024-03-15' suffix]",
			want: Query{
				Fuzzy:    "prefix[   suffix",
				Keywords: []string{"exact"},
				Flag:     Suffix,
				After:    carbon.CreateFromDate(2024, 3, 15),
			},
		},
		{
			name:     "invalid date after:",
			input:    "after:not-a-date",
			wantErrs: 1,
			want: Query{
				Fuzzy: "",
				Flag:  Fuzzy,
			},
		},
		{
			name:     "invalid date before:",
			input:    "before:garbage",
			wantErrs: 1,
			want: Query{
				Fuzzy: "",
				Flag:  Fuzzy,
			},
		},
		{
			name:  "unclosed quote (treated as fuzzy)",
			input: "'unclosed",
			want: Query{
				Fuzzy: "'unclosed",
				Flag:  Fuzzy,
			},
		},
		{
			name:  "relative date after:",
			input: "after:'yesterday'",
			want: Query{
				Fuzzy: "",
				Flag:  Fuzzy,
				After: carbon.Now().SubDay(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, warnings := Parse([]byte(tt.input))

			if len(warnings) != tt.wantErrs {
				t.Errorf("Parse() warnings count = %d, want %d", len(warnings), tt.wantErrs)
			}

			if got.Fuzzy != tt.want.Fuzzy {
				t.Errorf("Parse().Fuzzy = %v, want %v", got.Fuzzy, tt.want.Fuzzy)
			}
			if got.Regex != tt.want.Regex {
				t.Errorf("Parse().Regex = %v, want %v", got.Regex, tt.want.Regex)
			}
			if got.Flag != tt.want.Flag {
				t.Errorf("Parse().Flag = %v, want %v", got.Flag, tt.want.Flag)
			}
			if !reflect.DeepEqual(got.Keywords, tt.want.Keywords) {
				t.Errorf("Parse().Keywords = %v, want %v", got.Keywords, tt.want.Keywords)
			}

			if got.After.IsValid() != tt.want.After.IsValid() {
				t.Errorf("Parse().After.IsValid() = %v, want %v", got.After.IsValid(), tt.want.After.IsValid())
			}
			if got.After.IsValid() && got.After.DiffAbsInHours(tt.want.After) >= 12 {
				t.Errorf("Parse().After - <want> = %v, want <12", got.After.DiffAbsInHours(tt.want.After))
			}

			if got.Before.IsValid() != tt.want.Before.IsValid() {
				t.Errorf("Parse().Before.IsValid() = %v, want %v", got.Before.IsValid(), tt.want.Before.IsValid())
			}
			if got.Before.IsValid() && got.Before.DiffAbsInHours(tt.want.Before) >= 12 {
				t.Errorf("Parse().Before - <want> = %v, want <12", got.Before.DiffAbsInHours(tt.want.Before))
			}
		})
	}
}

func TestCollect(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		delim   byte
		want    string
		wantN   int
		wantEnd bool
	}{
		{
			name:    "simple quoted string",
			input:   "hello'",
			delim:   '\'',
			want:    "hello",
			wantN:   6, // "hello'"
			wantEnd: true,
		},
		{
			name:    "unclosed quote",
			input:   "hello",
			delim:   '\'',
			want:    "hello",
			wantN:   5,
			wantEnd: false,
		},
		{
			name:    "escaped quote",
			input:   "it\\'s done'",
			delim:   '\'',
			want:    "it's done",
			wantN:   11,
			wantEnd: true,
		},
		{
			name:    "empty string with delimiter",
			input:   "'",
			delim:   '\'',
			want:    "",
			wantN:   1,
			wantEnd: true,
		},
		{
			name:    "empty input",
			input:   "",
			delim:   '\'',
			want:    "",
			wantN:   0,
			wantEnd: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			n, ended := collect(buf, []byte(tt.input), tt.delim)

			if got := buf.String(); got != tt.want {
				t.Errorf("collect() string = %q, want %q", got, tt.want)
			}
			if n != tt.wantN {
				t.Errorf("collect() n = %d, want %d", n, tt.wantN)
			}
			if ended != tt.wantEnd {
				t.Errorf("collect() ended = %v, want %v", ended, tt.wantEnd)
			}
		})
	}
}
