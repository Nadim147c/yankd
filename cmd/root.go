package cmd

import (
	"context"
	"os"
	"path/filepath"

	"github.com/Nadim147c/fang"
	"github.com/adrg/xdg"
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	Command.PersistentFlags().
		StringP("database", "d", "XDG_DATA_HOME/yankd", "set database location directory")

	viper.SetEnvPrefix("yankd")
	viper.AutomaticEnv()

	carapace.Gen(Command)
}

// Command is the root command for yankd
var Command = &cobra.Command{
	Use:   "yankd",
	Short: "A dead simple clipboard manager",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		dbPath := filepath.Join(xdg.DataHome, "yankd")
		viper.SetDefault("database", dbPath)
		return nil
	},
}

// Execute runs the cobra cli
func Execute(version string) {
	err := fang.Execute(
		context.Background(),
		Command,
		fang.WithFlagTypes(),
		fang.WithShorthandPadding(),
		fang.WithVersion(version),
		fang.WithoutCompletions(),
	)
	if err != nil {
		os.Exit(1)
	}
}
