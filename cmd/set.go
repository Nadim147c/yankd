package cmd

import (
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/Nadim147c/yankd/internal/db"
	"github.com/spf13/cobra"
)

func init() {
	Command.AddCommand(setCommand)
}

// FIXME: please fix me......

var setCommand = &cobra.Command{
	Use:   "set",
	Short: "Set content of given id to clipboard",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		clip, err := db.Get(cmd.Context(), uint(id))
		if err != nil {
			return err
		}

		// TODO: set native protocal to set clipboard
		if clip.BlobPath == "" {
			slog.Debug("setting text content", "size", len(clip.Text))
			wlCopy := exec.Command("wl-copy")
			wlCopy.Stdin = strings.NewReader(clip.Text)
			return wlCopy.Run()
		}

		slog.Debug("setting binary content", "size", len(clip.Text))
		file, err := os.Open(clip.BlobPath)
		if err != nil {
			return err
		}
		defer file.Close()

		wlCopy := exec.Command("wl-copy")
		wlCopy.Stdin = file

		return wlCopy.Run()
	},
}
