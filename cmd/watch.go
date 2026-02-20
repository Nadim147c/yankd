package cmd

import (
	"context"
	"log/slog"
	"sync"

	"github.com/Nadim147c/yankd/internal/clipboard"
	"github.com/Nadim147c/yankd/internal/db"
	"github.com/Nadim147c/yankd/internal/ipc"
	"github.com/Nadim147c/yankd/internal/models"
	"github.com/spf13/cobra"
)

func init() {
	Command.AddCommand(watchCommand)
}

var watchCommand = &cobra.Command{
	Use:     "daemon",
	Aliases: []string{"watch"},
	Short:   "Start yankd deamon",
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()

		clips := make(chan models.ClipboardEvent)
		context.AfterFunc(ctx, func() { close(clips) })

		db, err := db.CreateDB()
		if err != nil {
			return err
		}
		defer db.Close()

		wlClient := clipboard.NewClient()
		defer wlClient.Close()

		var wg sync.WaitGroup

		ipcServer := ipc.NewServer(db, wlClient)

		wg.Go(func() { wlClient.Listen(ctx, clips) }) //nolint
		wg.Go(func() { ipcServer.Listen(ctx) })       //nolint
		wg.Go(func() {
			for clip := range clips {
				slog.Debug("Saving content to clipboard history", "mime", clip.PrimaryMimeType)
				if err := db.Insert(ctx, clip); err != nil {
					slog.Error("failed to insert data", "error", err)
				}
			}
		})

		wg.Wait()
		return nil
	},
}
