package cmd

import (
	"context"
	"log/slog"

	"github.com/Nadim147c/yankd/internal/db"
	"github.com/Nadim147c/yankd/pkg/clipboard"
	"github.com/spf13/cobra"
)

func init() {
	Command.AddCommand(watchCommand)
}

var watchCommand = &cobra.Command{
	Use:   "watch",
	Short: "Watch for clipboard changes",
	RunE: func(cmd *cobra.Command, _ []string) error {
		slog.Info("yankd watch starting", "version", Command.Version)
		ctx := cmd.Context()

		clips := make(chan clipboard.Clip)
		context.AfterFunc(ctx, func() { close(clips) })

		go clipboard.Watch(ctx, clips)

		if err := db.InitializeFTS(); err != nil {
			return err
		}
		defer db.Close()

		for clip := range clips {
			slog.Debug("Saving content to clipboard history", "mime", clip.Mime)
			db.Insert(ctx, clip)
		}
		return ctx.Err()
	},
}
