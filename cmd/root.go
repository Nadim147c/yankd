package cmd

import (
	"context"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/charmbracelet/log"

	"github.com/Nadim147c/fang"
	"github.com/adrg/xdg"
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	pfset := Command.PersistentFlags()
	pfset.StringP("database", "d", "XDG_DATA_HOME/yankd", "set database location directory")
	pfset.CountP("verbose", "v", "set log level")
	pfset.BoolP("quiet", "q", false, "suppress all the logs")

	viper.SetEnvPrefix("yankd")
	viper.AutomaticEnv()

	carapace.Gen(Command)
}

// Command is the root command for yankd
var Command = &cobra.Command{
	Use:   "yankd",
	Short: "A dead simple clipboard manager",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		viper.BindPFlags(cmd.Flags())

		level := log.ErrorLevel - (log.Level(viper.GetInt("verbose") * 4))
		if viper.GetBool("quiet") {
			level = math.MaxInt
		}

		logger := log.NewWithOptions(os.Stderr, log.Options{
			TimeFormat: time.RFC822,
			Level:      level,
		})

		slog.SetDefault(slog.New(logger))

		dbPath := filepath.Join(xdg.DataHome, "yankd")
		viper.SetDefault("database", dbPath)

		slog.Info("Logger is has been setup", "level", level)

		return nil
	},
}

// Execute runs the cobra cli
func Execute(version string) {
	err := fang.Execute(
		context.Background(),
		Command,
		fang.WithNotifySignal(syscall.SIGINT, syscall.SIGTERM),
		fang.WithFlagTypes(),
		fang.WithShorthandPadding(),
		fang.WithVersion(version),
		fang.WithoutCompletions(),
	)
	if err != nil {
		os.Exit(1)
	}
}
