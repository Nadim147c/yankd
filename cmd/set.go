package cmd

import (
	"log/slog"

	"github.com/Nadim147c/yankd/internal/ipc"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	Command.AddCommand(setCommand)
}

var setCommand = &cobra.Command{
	Use:   "set",
	Short: "Set clipboard to given id",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		return viper.BindPFlags(cmd.Flags())
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := uuid.Parse(args[0])
		if err != nil {
			return err
		}

		res, err := ipc.SetEvent(cmd.Context(), id)
		if err != nil {
			return err
		}
		mimes := make([]string, len(res.MimeType))
		for i, entry := range res.Entries {
			mimes[i] = entry.MimeType
		}
		slog.Info("Clipboard updated", "id", res.ID, "mime-type", mimes)
		return nil
	},
}
