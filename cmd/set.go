package cmd

import (
	"strconv"

	"github.com/Nadim147c/yankd/internal/db"
	"github.com/Nadim147c/yankd/internal/ipc"
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
		db, err := db.CreateDB()
		if err != nil {
			return err
		}
		defer db.Close()

		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return err
		}

		ipcClient := ipc.NewClient()
		return ipcClient.SendSet(cmd.Context(), id)
	},
}
