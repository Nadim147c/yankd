package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/Nadim147c/yankd/internal/ipc"
	"github.com/gabriel-vasile/mimetype"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	fileInputs FileMapValue
	dataInputs DataMapValue
)

func init() {
	copyCommand.Flags().VarP(&fileInputs, "file", "f", "Input clipboard content from file")
	copyCommand.Flags().VarP(&dataInputs, "data", "D", "Input clipboard content from value")
	Command.AddCommand(copyCommand)
}

type dataMap map[string][]byte

func (d dataMap) setNoOverride(key string, val []byte) error {
	if _, ok := d[key]; ok {
		return fmt.Errorf("mimetype %s used more than once", key)
	}
	d[key] = val
	return nil
}

func readFileToBuf(buf *bytes.Buffer, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = buf.ReadFrom(file)
	return err
}

var copyCommand = &cobra.Command{
	Use:   "copy",
	Short: "Copy items of one or more mime-type.",
	Example: `
  # Copy a single file (auto-detects MIME type)
  yankd copy /path/to/image.png

  # Copy multiple files at once
  yankd copy /path/to/file1.txt /path/to/file2.jpg /path/to/file3.html

  # Pipe content from standard input (auto-detects MIME type)
  echo "hello world" | yankd copy

  # Copy file content with an explicit MIME type mapping
  yankd copy --file text/html=/path/to/index.html

  # Copy raw text or data directly via CLI flag
  yankd copy --data text/plain="Hello World"

  # Combine flags and positional files into a single clipboard payload
  yankd copy \
    --data text/plain="Some raw text" \
    --file text/html=/path/to/index.html \
    /path/to/fallback.txt`,
	PreRunE: func(cmd *cobra.Command, _ []string) error {
		return viper.BindPFlags(cmd.Flags())
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		payload := dataMap{}
		maps.Copy(payload, dataInputs)

		buf := bytes.NewBuffer(nil)

		for mime, path := range fileInputs {
			buf.Reset()
			err := readFileToBuf(buf, path)
			if err != nil {
				return err
			}

			err = payload.setNoOverride(mime, buf.Bytes())
			if err != nil {
				return err
			}
		}

		stat, _ := os.Stdin.Stat()
		if stat.Mode()&os.ModeNamedPipe != 0 {
			buf.Reset()
			_, err := buf.ReadFrom(os.Stdin)
			if err != nil {
				return err
			}

			mt := mimetype.Detect(buf.Bytes())
			err = payload.setNoOverride(mt.String(), buf.Bytes())
			if err != nil {
				return err
			}
		}

		for _, file := range args {
			buf.Reset()
			err := readFileToBuf(buf, file)
			if err != nil {
				return err
			}

			mt := mimetype.Detect(buf.Bytes())
			err = payload.setNoOverride(mt.String(), buf.Bytes())
			if err != nil {
				return err
			}
		}

		mimes := slices.Collect(maps.Keys(payload))
		slog.Info("sending clipboard request", "mimetypes", mimes)

		event, err := ipc.SetClipboard(cmd.Context(), map[string][]byte(payload))
		if err != nil {
			return err
		}

		slog.Info("clipboard set", "database-id", event.ID)
		return nil
	},
}

// FileMapValue implements pflag.Value for map[string]string flags.
type FileMapValue map[string]string

func (f *FileMapValue) String() string {
	return "empty"
}

func (f *FileMapValue) Set(value string) error {
	k, v, ok := strings.Cut(value, "=")
	if !ok {
		return errors.New("invalid value format")
	}

	if *f == nil {
		*f = make(map[string]string)
	}

	if _, exists := (*f)[k]; exists {
		return fmt.Errorf("duplicate type %q provided", k)
	}

	(*f)[k] = v
	return nil
}

func (f *FileMapValue) Type() string {
	return "type=path"
}

// DataMapValue implements pflag.Value for map[string][]byte flags.
type DataMapValue map[string][]byte

func (d *DataMapValue) String() string {
	return "empty"
}

func (d *DataMapValue) Set(value string) error {
	k, v, ok := strings.Cut(value, "=")
	if !ok {
		return errors.New("invalid value format")
	}

	if *d == nil {
		*d = make(map[string][]byte)
	}

	if _, exists := (*d)[k]; exists {
		return fmt.Errorf("duplicate type %q provided", k)
	}

	(*d)[k] = []byte(v)
	return nil
}

func (d *DataMapValue) Type() string {
	return "type=data"
}
