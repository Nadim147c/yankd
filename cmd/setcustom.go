package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Nadim147c/yankd/internal/clipboard"
	"github.com/Nadim147c/yankd/internal/ipc"
	"github.com/Nadim147c/yankd/internal/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	Command.AddCommand(setCustomCommand)
	f := setCustomCommand.Flags()
	f.StringP("mime", "m", "", "MIME type for single entry (reads raw bytes from stdin)")
	f.Bool("json", false, "Read JSON with multiple entries from stdin")
	setCustomCommand.MarkFlagsMutuallyExclusive("mime", "json")
}

var setCustomCommand = &cobra.Command{
	Use:   "set-custom",
	Short: "Set custom clipboard content",
	Long: `Set arbitrary clipboard content with one or more MIME types.

Single MIME (raw bytes from stdin):
  yankd set-custom -m text/plain < file.txt
  echo "hello" | yankd set-custom -m text/plain

Multiple MIME (JSON with base64 blobs from stdin):
  echo '{"entries":[
    {"mime_type":"text/plain","blob":"aGVsbG8="},
    {"mime_type":"text/html","blob":"PGI+aGVsbG88L2I+"}
  ]}' | yankd set-custom --json
`,
	Args: cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		return viper.BindPFlags(cmd.Flags())
	},
	RunE: func(cmd *cobra.Command, _ []string) error {
		mime := viper.GetString("mime")
		jsonMode := viper.GetBool("json")

		if mime == "" && !jsonMode {
			return errors.New("use --mime for single entry or --json for multi-entry")
		}

		var event models.ClipboardEvent
		event.Time = time.Now()

		if mime != "" {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}
			event.Entries = []models.ClipboardEntry{{
				MimeType: mime,
				Blob:     data,
				IsText:   utf8.Valid(data),
				Hash:     models.Hash{}, // zero — JSON marshal/unmarshal corrupts Hash type, daemon recomputes from blob
			}}
		} else {
			var input struct {
				Entries []struct {
					MimeType string `json:"mime_type"`
					Blob     []byte `json:"blob"`
				} `json:"entries"`
			}
			if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
				return fmt.Errorf("failed to parse JSON: %w", err)
			}
			if len(input.Entries) == 0 {
				return errors.New("JSON must contain at least one entry")
			}
			for _, e := range input.Entries {
				if e.MimeType == "" {
					continue
				}
				event.Entries = append(event.Entries, models.ClipboardEntry{
					MimeType: e.MimeType,
					Blob:     e.Blob,
					IsText:   utf8.Valid(e.Blob),
					Hash:     models.Hash{}, // zero — JSON marshal/unmarshal corrupts Hash type, daemon recomputes from blob
				})
			}
			if len(event.Entries) == 0 {
				return errors.New("no valid entries found in JSON")
			}
		}

		// Set primary MIME type (prefer image/*, fallback to first)
		event.MimeType = event.Entries[0].MimeType
		for _, e := range event.Entries {
			if strings.HasPrefix(e.MimeType, "image/") {
				event.MimeType = e.MimeType
				break
			}
		}

		event.Preview = clipboard.GeneratePreview(event.Entries)

		if err := ipc.SetCustom(cmd.Context(), event); err != nil {
			return fmt.Errorf("failed to set clipboard: %w", err)
		}

		mimes := make([]string, len(event.Entries))
		for i, e := range event.Entries {
			mimes[i] = e.MimeType
		}
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Clipboard set:", strings.Join(mimes, ", "))
		return nil
	},
}
