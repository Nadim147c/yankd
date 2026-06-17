package cmd

import (
	"log/slog"

	"github.com/Nadim147c/yankd/internal/ipc"
	"github.com/google/uuid"
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
  yankd delete 019eda41-3efb-7f0c-a7f9-df9920000f3d

  # Delete multiple items with IDs 1, 5, and 10
  yankd delete 019eda41-5a07-7a12-b3b7-1efb7553eb89 019eda41-6d69-74ad-82db-f1ac4bda9b5b
  `,
	Args: cobra.MinimumNArgs(1),
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		return viper.BindPFlags(cmd.Flags())
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ids := make([]uuid.UUID, len(args))
		for i, arg := range args {
			id, err := uuid.Parse(arg)
			if err != nil {
				return err
			}
			ids[i] = id
		}

		n, err := ipc.DeteteEvents(cmd.Context(), ids...)
		if err != nil {
			return err
		}
		slog.Info("Clipboard history deleted", "deleted-items", n)
		return nil
	},
}
