package clipboard

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/Nadim147c/yankd/internal/models"
)

// TestGeneratePreivewKeepsPreviewValidUTF8 covers the byte cap: the preview
// buffer is cut at maxPreviewLength, and that cut must not land inside a
// multi-byte rune. An invalid UTF-8 preview is rejected by DuckDB when it is
// bound as a VARCHAR, which loses the whole clipboard event.
func TestGeneratePreivewKeepsPreviewValidUTF8(t *testing.T) {
	cases := []struct {
		name string
		blob string
	}{
		{"three byte runes", strings.Repeat("中", 3000)},      // 9000 bytes, cut lands mid-rune
		{"four byte runes", "a" + strings.Repeat("🙂", 2250)}, // 9001 bytes, cut lands mid-rune
		{"four byte runes aligned", strings.Repeat("🙂", 2250)},
		{"ascii", strings.Repeat("a", 9000)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := GeneratePreivew([]models.ClipboardEntry{{
				MimeType: "text/plain",
				Blob:     []byte(tc.blob),
				IsText:   true,
			}})

			if !utf8.ValidString(got) {
				t.Errorf("preview is not valid UTF-8 (tail %q)", got[len(got)-4:])
			}
			if len(got) > maxPreviewLength {
				t.Errorf("preview is %d bytes, want at most %d", len(got), maxPreviewLength)
			}
			// A rune-safe cut costs at most the partial rune it backs off from.
			if len(got) < maxPreviewLength-3 {
				t.Errorf("preview is %d bytes, want at least %d", len(got), maxPreviewLength-3)
			}
		})
	}
}
