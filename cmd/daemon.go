package cmd

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/Nadim147c/yankd/internal/clipboard"
	"github.com/Nadim147c/yankd/internal/db"
	"github.com/Nadim147c/yankd/internal/ipc"
	"github.com/Nadim147c/yankd/internal/models"
	"github.com/spf13/cobra"
)

func init() {
	Command.AddCommand(daemonCommand)
}

var daemonCommand = &cobra.Command{
	Use:     "daemon",
	Aliases: []string{"watch"},
	Short:   "Start yankd daemon",
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		clips := make(chan models.ClipboardEvent)
		context.AfterFunc(ctx, func() {
			close(clips)
		})

		db, err := db.CreateDB()
		if err != nil {
			return err
		}
		defer db.Close()

		wlClient := clipboard.NewClient()
		defer wlClient.Close()

		var wg sync.WaitGroup

		ipcServer := ipc.NewServer(db, wlClient)

		wg.Go(func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					n, err := db.DeleteDuplicates(ctx)
					if err != nil {
						slog.Error("failed to delete duplicates", "error", err)
						continue
					}
					slog.Info("remove duplicates events", "affected", n)
				}
			}
		})

		wg.Go(func() {
			err := wlClient.Listen(ctx, clips)
			if err != nil && !errors.Is(err, context.Canceled) {
				slog.Error("failed to start wayland client", "error", err)
				cancel()
			}
		})

		wg.Go(func() {
			err := ipcServer.Listen(ctx)
			if err != nil && !errors.Is(err, context.Canceled) {
				slog.Error("failed to start ipc client", "error", err)
				cancel()
			}
		})

		wg.Go(func() {
			for clip := range clips {
				slog.Debug("Saving content to clipboard history", "mime", clip.MimeType)
				if err := db.Insert(ctx, &clip); err != nil {
					slog.Error("failed to insert data", "error", err)
				}
			}
		})

		wg.Wait()
		return nil
	},
}
