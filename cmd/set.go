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
	setCommand.Flags().BoolP("promote", "p", false, "Update timestamp to promote item to top")
}

var setCommand = &cobra.Command{
	Use:   "set",
	Short: "Set clipboard to given id",
	Long: `Set clipboard content from history by event ID.
With --promote, the event timestamp is updated to now,
moving it to the top of the list on next load.`,
	Example: `
    # Restore clipboard of given id
    yankd set 019fb267-e9b8-700d-a100-643a7e8bf651
    # Restore clipboard of given id and update timestamp
    yankd set --promote 019fb267-e9b8-700d-a100-643a7e8bf651
  `,
	Args: cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		return viper.BindPFlags(cmd.Flags())
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := uuid.Parse(args[0])
		if err != nil {
			return err
		}

		res, err := ipc.SetEvent(cmd.Context(), id, viper.GetBool("promote"))
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
