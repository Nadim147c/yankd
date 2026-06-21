package cmd

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"slices"

	"github.com/Nadim147c/yankd/internal/ipc"
	"github.com/Nadim147c/yankd/internal/models"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	Command.AddCommand(getCommand)
	getCommand.Flags().BoolP("primary", "p", false, "Return data for primary mime type")
	getCommand.Flags().BoolP("base64", "b", false, "Return binary data for primary mime type as base64")
}

var getCommand = &cobra.Command{
	Use:   "get",
	Short: "Get content of given id from history",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		return viper.BindPFlags(cmd.Flags())
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := uuid.Parse(args[0])
		if err != nil {
			return err
		}

		event, err := ipc.GetEvent(cmd.Context(), id)
		if err != nil {
			return err
		}

		if viper.GetBool("primary") {
			idx := slices.IndexFunc(event.Entries, func(e models.ClipboardEntry) bool {
				return e.MimeType == event.MimeType
			})

			if idx < 0 {
				return errors.New("primary mime_type not found")
			}

			entry := event.Entries[idx]
			if entry.IsText {
				_, err := os.Stdout.Write(entry.Blob)
				return err
			}

			if viper.GetBool("base64") {
				w := base64.NewEncoder(base64.StdEncoding, os.Stdout)
				if _, err := w.Write(entry.Blob); err != nil {
					return err
				}
				return w.Close()
			}

			_, err := os.Stdout.Write(entry.Blob)
			return err
		}

		return json.NewEncoder(cmd.OutOrStdout()).Encode(event)
	},
}
