package clipboard

import (
	"bytes"
	"errors"
	"os"
	"testing"
	"time"
)

// TestReadWithIdleTimeout covers the two properties the watcher depends on: a
// slow but continuous source is read to the end, and a source that goes silent
// is abandoned instead of blocking forever.
func TestReadWithIdleTimeout(t *testing.T) {
	const (
		idle = 100 * time.Millisecond
		gap  = 40 * time.Millisecond
	)

	t.Run("keeps reading a source that never goes idle", func(t *testing.T) {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		defer r.Close() //nolint

		chunks := [][]byte{
			bytes.Repeat([]byte("a"), 4096),
			bytes.Repeat([]byte("b"), 4096),
			bytes.Repeat([]byte("c"), 4096),
			bytes.Repeat([]byte("d"), 4096),
		}
		go func() {
			defer w.Close() //nolint
			for _, chunk := range chunks {
				// Every gap stays below the idle timeout, the total does not.
				time.Sleep(gap)
				if _, err := w.Write(chunk); err != nil {
					return
				}
			}
		}()

		start := time.Now()
		got, err := readWithIdleTimeout(r, idle)
		if err != nil {
			t.Fatalf("readWithIdleTimeout() error = %v", err)
		}

		want := bytes.Join(chunks, nil)
		if !bytes.Equal(got, want) {
			t.Errorf("read %d bytes, want %d", len(got), len(want))
		}
		if elapsed := time.Since(start); elapsed <= idle {
			t.Errorf("read took %v, want longer than the idle timeout %v", elapsed, idle)
		}
	})

	t.Run("gives up when the source goes silent", func(t *testing.T) {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		defer r.Close() //nolint
		defer w.Close() //nolint

		start := time.Now()
		got, err := readWithIdleTimeout(r, idle)
		if !errors.Is(err, os.ErrDeadlineExceeded) {
			t.Fatalf("readWithIdleTimeout() error = %v, want a deadline error", err)
		}
		if len(got) != 0 {
			t.Errorf("read %d bytes, want none", len(got))
		}
		if elapsed := time.Since(start); elapsed > time.Second {
			t.Errorf("gave up after %v, want about %v", elapsed, idle)
		}
	})

	t.Run("reports partial data as an error", func(t *testing.T) {
		r, w, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		defer r.Close() //nolint
		defer w.Close() //nolint

		if _, err := w.Write([]byte("partial")); err != nil {
			t.Fatal(err)
		}

		got, err := readWithIdleTimeout(r, idle)
		if err == nil {
			t.Fatalf("readWithIdleTimeout() error = nil, want a deadline error after %q", got)
		}
		if string(got) != "partial" {
			t.Errorf("read %q, want %q", got, "partial")
		}
	})
}
