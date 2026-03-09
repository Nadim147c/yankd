package clipboard

import (
	"bytes"
	"regexp"
	"slices"
	"strings"

	"github.com/Nadim147c/yankd/internal/models"
	"golang.org/x/net/html"
)

const maxPreviewLength = 1024 * 8 // 8KB

var htmlTagsSortScore = []string{"alt", "content", "url", "src", "title", "href"}

type wordWritter struct {
	buf     *bytes.Buffer
	done    bool
	written map[string]struct{}
}

func (w *wordWritter) WriteWords(words ...string) (done bool) {
	if w.done {
		return true
	}
	for _, word := range words {
		if _, ok := w.written[word]; ok {
			continue
		}
		w.written[word] = struct{}{}
		w.buf.Grow(len(word) + 1)
		w.buf.WriteString(word)
		w.buf.WriteByte(' ')
		if w.buf.Len() > maxPreviewLength {
			w.buf.Truncate(maxPreviewLength)
			w.done = true
			return true
		}
	}
	return false
}

func (w *wordWritter) String() string {
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(w.buf.String(), " ")
}

func generatePreivew(entires []models.ClipboardEntry) string {
	buf := &wordWritter{
		buf:     bytes.NewBuffer(make([]byte, 0, maxPreviewLength)),
		written: make(map[string]struct{}),
	}

	sortedEntries := slices.Clone(entires)
	slices.SortFunc(sortedEntries, func(a, b models.ClipboardEntry) int {
		if a.MimeType == "text/html" {
			return 1
		}
		if b.MimeType == "text/html" {
			return -1
		}
		return strings.Compare(a.MimeType, b.MimeType)
	})

	for _, entry := range sortedEntries {
		if !entry.IsText {
			continue
		}

		if entry.MimeType == "text/html" {
			m := parseHtml(entry.Blob)
			for _, tag := range htmlTagsSortScore {
				if val, ok := m[tag]; ok {
					if buf.WriteWords(val) {
						return buf.String()
					}
					delete(m, tag)
				}

				for k, v := range m {
					if buf.WriteWords(k, v) {
						return buf.String()
					}
				}
			}
			continue
		}

		if entry.IsText || len(entry.Blob) > 0 {
			if buf.WriteWords(string(entry.Blob)) {
				return buf.String()
			}
		}
	}

	return buf.String()
}

func parseHtml(p []byte) map[string]string {
	doc, err := html.Parse(bytes.NewReader(p))
	if err != nil {
		return nil
	}

	m := map[string]string{}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, attr := range n.Attr {
				k := strings.TrimSpace(attr.Key)
				v := strings.TrimSpace(attr.Val)
				if k != "" && v != "" {
					m[k] = v
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return m
}
