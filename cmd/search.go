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

		limit := viper.GetInt64("limit")

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
	simpleEvents := make([]SimplifedEvent, len(events))
	for i, event := range events {
		simpleEvents[i] = SimplifedEvent{
			ID:       event.ID,
			MimeType: event.PrimaryMimeType,
			Time:     event.Time,
			Preview:  event.Preview,
		}
	}
	return json.NewEncoder(os.Stdout).Encode(simpleEvents)
}

func formatJSON(events []models.ClipboardEvent) error {
	return json.NewEncoder(os.Stdout).Encode(events)
}

func formatPlain(events []models.ClipboardEvent) error { //nolint:unparam
	for event := range slices.Values(events) {
		fmt.Fprintln(os.Stdout, event.ID, event.PrimaryMimeType, event.Preview) //nolint:errcheck
	}
	return nil
}
