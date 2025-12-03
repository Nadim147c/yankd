package cmd

import (
	"log/slog"

	"github.com/Nadim147c/yankd/internal/db"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	Command.AddCommand(wipeCommand)
}

var wipeCommand = &cobra.Command{
	Use:   "wipe",
	Short: "Delete all clipboard history",
	Args:  cobra.NoArgs,
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		return viper.BindPFlags(cmd.Flags())
	},
	RunE: func(cmd *cobra.Command, _ []string) error {
		n, err := db.Wipe(cmd.Context())
		if err != nil {
			return err
		}
		slog.Info("Clipboard history deleted", "deleted-items", n)
		return nil
	},
}
