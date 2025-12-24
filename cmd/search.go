package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/Nadim147c/yankd/internal/db"
	"github.com/Nadim147c/yankd/pkg/clipboard"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	Command.AddCommand(searchCommand)
	fset := searchCommand.Flags()
	fset.BoolP("sync", "s", false, "synchronize database before search")
	fset.IntP("limit", "n", 40, "number of items to display")
	fset.StringP(
		"format", "f", "simple",
		"output format (simple, json, json-stream, or Go template string)",
	)
}

var searchCommand = &cobra.Command{
	Use:   "search <query>",
	Short: "Search clipboard history",
	Long:  "Search through clipboard history for items matching the query",
	Example: `
  # Search for "password" in clipboard history
  yankd search password

  # Limit results to 10 items in JSON format
  yankd search password --limit 10 --format json

  # Sync database before searching
  yankd search password --sync

  # Use custom template
  yankd search password --format "{{.ID}}: {{.Text}}"
  `,
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		viper.SetDefault("limit", 40)
		return viper.BindPFlags(cmd.Flags())
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")

		sync := viper.GetBool("sync")
		limit := viper.GetInt("limit")

		clips, err := db.Search(cmd.Context(), query, limit, sync)
		if err != nil {
			return err
		}
		defer db.Close()

		format := viper.GetString("format")
		switch strings.ToLower(format) {
		case "simple":
			return formatSimple(clips)
		case "json":
			return formatJSON(clips)
		case "json-stream":
			return formatJSONStream(clips)
		default:
			return formatTemplate(clips, format)
		}
	},
}

// formatSimple outputs results in a simple tab-separated format
func formatSimple(clips []clipboard.Clip) error {
	for _, clip := range clips {
		fmt.Printf("%d\t%s\t%s\n", clip.ID, clip.Mime, simpleClip(clip))
	}
	return nil
}

// formatJSON outputs all results as a single JSON array
func formatJSON(clips []clipboard.Clip) error {
	return json.NewEncoder(os.Stdout).Encode(clips)
}

// formatJSONStream outputs each result as a single-line JSON object
func formatJSONStream(clips []clipboard.Clip) error {
	encoder := json.NewEncoder(os.Stdout)
	for _, clip := range clips {
		if err := encoder.Encode(clip); err != nil {
			return err
		}
	}
	return nil
}

var templateFunc = template.FuncMap{
	"simplify": simpleText,
	"fallback": fallbackText,
}

func simpleText(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// fallbackText returns the first non-empty string from the provided values
func fallbackText(values ...string) string {
	for _, val := range values {
		if val != "" {
			return val
		}
	}
	return ""
}

// formatTemplate outputs results using a Go template string
func formatTemplate(clips []clipboard.Clip, tmplStr string) error {
	tmpl, err := template.New("search").Funcs(templateFunc).Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("invalid template: %w", err)
	}
	for _, clip := range clips {
		if err := tmpl.Execute(os.Stdout, clip); err != nil {
			return err
		}
	}
	return nil
}

// simpleClip extracts and formats text from a clipboard item
func simpleClip(clip clipboard.Clip) string {
	text := fallbackText(clip.Text, clip.BlobPath, clip.Metadata)
	out := simpleText(text)
	if len(out) > 100 {
		out = out[:100]
	}
	return out
}
