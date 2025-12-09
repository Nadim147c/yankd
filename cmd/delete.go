package cmd

import (
	"log/slog"

	"github.com/Nadim147c/yankd/internal/db"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	Command.AddCommand(deleteCommand)
}

var deleteCommand = &cobra.Command{
	Use:   "delete ...ids",
	Short: "Remove items from clipboard history",
	Example: `
  # Delete a single item with ID 42
  yankd delete 42

  # Delete multiple items with IDs 1, 5, and 10
  yankd delete 1 5 10

  # Delete a range of items (using shell expansion)
  yankd delete {20..25}

  # Delete sensitive items like PEM encoded keys
  yankd search --limit 10000 "BEGIN KEY" | awk '{ print $1 }' | xargs yankd delete
  `,
	Args: cobra.MinimumNArgs(1),
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		return viper.BindPFlags(cmd.Flags())
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ids, err := cast.ToUintSliceE(args)
		if err != nil {
			return err
		}
		n, err := db.Delete(cmd.Context(), ids)
		if err != nil {
			return err
		}
		slog.Info("Clipboard history deleted", "deleted-items", n)
		return nil
	},
}
