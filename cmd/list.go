package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/Nadim147c/yankd/internal/db"
	"github.com/Nadim147c/yankd/internal/db/sqlc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	Command.AddCommand(listCommand)
	flags := listCommand.Flags()
	flags.IntP("limit", "n", 100_000, "number of items to display")
	flags.StringP("format", "f", "", "number of items to display")
}

var listCommand = &cobra.Command{
	Use:   "list",
	Short: "List clipboard history",
	Long:  "List clipboard history",
	Example: `
  # List clipboard history (JSON format)
  yankd list --format=json --limit=100 | jq

  # Interactive search via fzf with preview (Binary data base64 encoded)
  yankd list -q | fzf -d\t --preview='yankd get -pqb {1}'
  `,
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		viper.SetDefault("limit", 100_000)
		return viper.BindPFlags(cmd.Flags())
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		limit := viper.GetInt64("limit")

		db, err := db.CreateDB()
		if err != nil {
			return err
		}
		defer db.Close()

		events, err := db.List(cmd.Context(), limit)
		if err != nil {
			return err
		}
		defer db.Close()

		switch viper.GetString("format") {
		case "json":
			return json.NewEncoder(os.Stdout).Encode(events)
		default:
			return formatPlainPreview(events)
		}
	},
}

func formatPlainPreview(events []sqlc.GetEventsPreviewAndIDRow) error { //nolint:unparam
	for event := range slices.Values(events) {
		fmt.Fprintf(os.Stdout, "%d\t%s\t%s\n", event.ID, event.PrimaryMimeType, event.Preview) //nolint:errcheck
	}
	return nil
}
