package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Nadim147c/yankd/internal/db"
	"github.com/Nadim147c/yankd/pkg/clipboard"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	Command.AddCommand(searchCommand)
	fset := searchCommand.Flags()
	fset.BoolP("sync", "s", false, "synchronize database before search")
	fset.IntP("limit", "n", 40, "number of item to list")
}

var searchCommand = &cobra.Command{
	Use:   "search",
	Short: "Search clipboard",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		viper.SetDefault("limit", 40)
		return viper.BindPFlags(cmd.Flags())
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		sync := viper.GetBool("sync")
		limit := viper.GetInt("limit")
		clips, err := db.Search(cmd.Context(), query, limit, sync)
		if err != nil {
			return err
		}
		for clip := range slices.Values(clips) {
			fmt.Printf("%d\t%s\n", clip.ID, textOrMeta(clip))
		}
		return nil
	},
}

func textOrMeta(clip clipboard.Clip) string {
	text := clip.Text
	if text == "" && clip.Metadata != "" {
		text = clip.Metadata
	}

	fields := strings.Fields(text)

	out := strings.Join(fields, " ")
	if len(out) > 100 {
		out = out[:100]
	}
	return out
}
