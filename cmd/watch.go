package cmd

import (
	"context"
	"log/slog"

	"github.com/Nadim147c/yankd/internal/clipboard"
	"github.com/Nadim147c/yankd/internal/db"
	"github.com/Nadim147c/yankd/internal/models"
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

		clips := make(chan models.Event)
		context.AfterFunc(ctx, func() { close(clips) })

		client := clipboard.NewClient()
		defer client.Close()

		go client.Watch(ctx, clips)

		db, err := db.CreateDB()
		if err != nil {
			return err
		}
		defer db.Close()

		for clip := range clips {
			slog.Debug("Saving content to clipboard history", "mime", clip.PrimaryMimeType)
			if err := db.Insert(ctx, clip); err != nil {
				slog.Error("failed to insert data", "error", err)
			}
		}
		return ctx.Err()
	},
}
