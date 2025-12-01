package cmd

import (
	"log/slog"

	"github.com/Nadim147c/yankd/internal/db"
	"github.com/Nadim147c/yankd/pkgs/clipboard"
	"github.com/spf13/cobra"
)

func init() {
	Command.AddCommand(Watch)
}

var Watch = &cobra.Command{
	Use:   "watch",
	Short: "Watch for clipboard changes and save the history",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		clips := make(chan clipboard.Clip)

		go func() {
			clipboard.Watch(ctx, clips)
			close(clips)
		}()

		for clip := range clips {
			slog.Debug("Saving content to clipboard history", "mime", clip.Mime)
			err := db.Insert(ctx, &clip)
			if err != nil {
				slog.Info("Failed to save clipboard", "error", err)
			}
		}
		return nil
	},
}
