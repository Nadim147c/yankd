package cmd

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/Nadim147c/yankd/internal/db"
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
			return json.NewEncoder(os.Stdout).Encode(events)
		default:
			return formatPlainPreview(events)
		}
	},
}
