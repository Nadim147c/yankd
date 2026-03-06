package cmd

import (
	"strconv"

	"github.com/Nadim147c/yankd/internal/ipc"
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
		client := ipc.NewClient()

		if len(args) == 0 || args[0] == "toggle" {
			return client.SendPause(ipc.PauseCmdToggle)
		}

		b, err := strconv.ParseBool(args[0])
		if err != nil {
			return err
		}

		newState := ipc.PauseCmdFalse
		if b {
			newState = ipc.PauseCmdTrue
		}

		return client.SendPause(newState)
	},
}
