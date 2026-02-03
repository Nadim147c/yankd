package cmd

import (
	"encoding/json"
	"errors"
	"os"
	"slices"

	"github.com/Nadim147c/yankd/internal/db"
	"github.com/Nadim147c/yankd/internal/models"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	Command.AddCommand(getCommand)
	getCommand.Flags().BoolP("primary", "p", false, "Return data from primary mime type")
}

var getCommand = &cobra.Command{
	Use:   "get",
	Short: "Get content of given id from history",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		return viper.BindPFlags(cmd.Flags())
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := cast.ToInt64E(args[0])
		if err != nil {
			return err
		}
		db, err := db.CreateDB()
		if err != nil {
			return err
		}
		defer db.Close()

		event, err := db.Get(cmd.Context(), id)
		if err != nil {
			return err
		}

		if viper.GetBool("primary") {
			idx := slices.IndexFunc(event.Entries, func(e models.Entry) bool {
				return e.MimeType == event.PrimaryMimeType
			})
			if idx < 0 {
				return errors.New("primary mime_type not found")
			}
			entry := event.Entries[idx]
			if entry.IsText {
				_, err := os.Stdout.WriteString(entry.Text.String)
				return err
			}
			_, err := os.Stdout.Write(entry.Blob)
			return err
		}

		return json.NewEncoder(cmd.OutOrStdout()).Encode(event)
	},
}
