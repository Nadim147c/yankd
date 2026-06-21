package cmd

import (
	"log/slog"

	"github.com/Nadim147c/yankd/internal/ipc"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

func init() {
	daemonCommand.AddCommand(daemonPauseCommand)
}

var daemonPauseCommand = &cobra.Command{
	Use:   "pause [toggle|true|false]",
	Short: "Pause history saving",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 || args[0] == "toggle" {
			v, err := ipc.SetPause(cmd.Context(), ipc.PauseToggle)
			if err != nil {
				return err
			}
			slog.Info("toggled history saving", "is_paused", v)
			return nil
		}

		toggleValue, err := cast.ToBoolE(args[0])
		if err != nil {
			return err
		}

		newState := ipc.PauseFalse
		if toggleValue {
			newState = ipc.PauseTrue
		}

		v, err := ipc.SetPause(cmd.Context(), newState)
		if err != nil {
			return err
		}
		slog.Info("toggled history saving", "is_paused", v)
		return nil
	},
}
