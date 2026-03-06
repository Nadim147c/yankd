package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/Nadim147c/yankd/internal/db"
	"github.com/Nadim147c/yankd/internal/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	Command.AddCommand(searchCommand)
	fset := searchCommand.Flags()
	fset.IntP("limit", "n", 40, "number of items to display")
	fset.StringP("format", "f", "", "number of items to display")
}

var searchCommand = &cobra.Command{
	Use:   "search <...query>",
	Short: "Search clipboard history",
	Long:  "Search through clipboard history for items matching the query",
	Example: `
  # Search for "password" in clipboard history
  yankd search password

  # Limit results to 10 items in JSON format
  yankd search password --limit 10

  # Sync database before searching
  yankd search password --sync
  `,
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		viper.SetDefault("limit", 40)
		return viper.BindPFlags(cmd.Flags())
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")

		limit := viper.GetInt("limit")

		db, err := db.CreateDB()
		if err != nil {
			return err
		}
		defer db.Close()

		events, err := db.Search(cmd.Context(), query, limit)
		if err != nil {
			return err
		}
		defer db.Close()

		switch viper.GetString("format") {
		case "json":
			return formatJSON(events)
		case "short-json":
			return formatSimpleJSON(events)
		default:
			return formatPlain(events)
		}
	},
}

type SimplifedEvent struct {
	ID       int64     `json:"id"`
	MimeType string    `json:"mime_type"`
	Time     time.Time `json:"time"`
	Preview  string    `json:"preview"`
}

func formatSimpleJSON(events []models.ClipboardEvent) error {
	simpleEvents := make([]SimplifedEvent, 0, len(events))
	for event := range slices.Values(events) {
		se := SimplifedEvent{
			ID:       event.ID,
			MimeType: event.PrimaryMimeType,
			Time:     event.Time,
			Preview:  getPreview(event.Entries),
		}
		simpleEvents = append(simpleEvents, se)
	}
	return json.NewEncoder(os.Stdout).Encode(simpleEvents)
}

func formatJSON(events []models.ClipboardEvent) error {
	return json.NewEncoder(os.Stdout).Encode(events)
}

func getPreview(entries []models.ClipboardEntry) string {
	m := make(map[models.Hash]struct{}, len(entries))
	uniqueEntries := slices.DeleteFunc(entries, func(e models.ClipboardEntry) bool {
		if !e.IsText || !e.Text.Valid {
			return true
		}
		if _, ok := m[e.Hash]; ok {
			return true
		}
		m[e.Hash] = struct{}{}
		return false
	})

	words := make([]string, 0, len(uniqueEntries))
	for entry := range slices.Values(uniqueEntries) {
		split := strings.Fields(entry.Text.String)
		words = append(words, split...)
	}

	if len(words) == 0 {
		return "<unknow clipboard>"
	}

	return strings.Join(words, " ")
}

func formatPlain(events []models.ClipboardEvent) error { //nolint:unparam
	for event := range slices.Values(events) {
		preview := fmt.Sprintf("%d\t%s\t%s", event.ID, event.PrimaryMimeType, getPreview(event.Entries))
		_, _ = fmt.Fprintln(os.Stdout, preview)
	}
	return nil
}
