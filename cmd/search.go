package cmd

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Nadim147c/yankd/internal/db"
	"github.com/Nadim147c/yankd/pkgs/clipboard"
	"github.com/spf13/cobra"
)

func init() {
	Command.AddCommand(Search)
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

var Search = &cobra.Command{
	Use:   "search",
	Short: "Search clipboard",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		clips, err := db.Search(cmd.Context(), query)
		if err != nil {
			return err
		}
		for clip := range slices.Values(clips) {
			fmt.Printf("%d\t%s\n", clip.ID, textOrMeta(clip))
		}
		return nil
	},
}
