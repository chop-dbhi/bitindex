package main

import (
	"compress/bzip2"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/chop-dbhi/bitindex"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func openFile(name string) (*os.File, io.Reader, error) {
	f, err := os.Open(name)

	if err != nil {
		return nil, nil, err
	}

	var r io.Reader

	// Detect compression.
	switch filepath.Ext(name) {
	case ".gzip", ".gz":
		r, err = gzip.NewReader(f)

		if err != nil {
			return nil, nil, err
		}

	case ".bzip2", ".bz2":
		r = bzip2.NewReader(f)

	default:
		r = f
	}

	return f, r, nil
}

var buildCmd = &cobra.Command{
	Use: "build [<path>]",

	Short: "Build an index.",

	Run: func(cmd *cobra.Command, args []string) {
		var (
			r   io.Reader
			err error
		)

		switch len(args) {
		case 0:
			r = os.Stdin

		case 1:
			var f *os.File

			if f, r, err = openFile(args[0]); err != nil {
				cmd.Printf("Cannot open file: %s\n", err)
				os.Exit(1)
			}

			defer f.Close()

		default:
			cmd.Println("Stdin or a single file must be passed.")
			os.Exit(1)
		}

		var ixer bitindex.Indexer

		switch viper.GetString("build.format") {
		case "csv":
			ix := bitindex.NewCSVIndexer(r)
			ix.Header = viper.GetBool("build.csv-header")

			kc := viper.GetInt("build.csv-key")
			dc := viper.GetInt("build.csv-domain")

			ix.Parse = func(row []string) (uint32, uint32, error) {
				ki, err := strconv.Atoi(row[kc])

				if err != nil {
					return 0, 0, err
				}

				di, err := strconv.Atoi(row[dc])

				if err != nil {
					return 0, 0, err
				}

				return uint32(ki), uint32(di), nil
			}

			ixer = ix

		default:
			cmd.Println("--format flag is required")
			os.Exit(1)
		}

		t0 := time.Now()
		idx, err := ixer.Index()
		dur := time.Now().Sub(t0)

		if err != nil {
			cmd.Printf("Error building index: %s\n", err)
			os.Exit(1)
		}

		cmd.Println("Build time:", dur)
		cmd.Println("Statistics:")
		cmd.Println("* Domain size:", idx.Domain.Size())
		cmd.Println("* Table size:", idx.Table.Size())
		cmd.Println("* Sparsity:", idx.Sparsity()*100)

		output := viper.GetString("build.output")

		var w io.Writer

		// Build output file.
		if output == "" {
			w = os.Stdout
		} else {
			o, err := os.Create(output)

			if err != nil {
				cmd.Println("Error opening file to write:", err)
				os.Exit(1)
			}

			w = o
		}

		if err := bitindex.DumpIndex(w, idx); err != nil {
			cmd.Println("Error dumping index:", err)
			os.Exit(1)
		}

		// Coerce to file, sync and close.
		if f, ok := w.(*os.File); ok {
			if err = f.Sync(); err != nil {
				cmd.Printf("Error syncing file: %s\n", err)
			}

			if err = f.Close(); err != nil {
				cmd.Printf("Error closing file: %s\n", err)
			}
		}
	},
}

func init() {
	flags := buildCmd.Flags()

	// General.
	flags.String("format", "", "Format of the input stream: csv")
	flags.String("output", "", "Specify an output file to write the stream to.")

	// format is required.
	buildCmd.MarkFlagRequired("format")

	viper.BindPFlag("build.format", flags.Lookup("format"))
	viper.BindPFlag("build.output", flags.Lookup("output"))

	// CSV indexer.
	flags.Bool("csv-header", false, "CSV file has a header")
	flags.Int("csv-key", 0, "Index of the column containing set keys.")
	flags.Int("csv-domain", 1, "Index of the column containing domain members.")

	viper.BindPFlag("build.csv-header", flags.Lookup("csv-header"))
	viper.BindPFlag("build.csv-key", flags.Lookup("csv-key"))
	viper.BindPFlag("build.csv-domain", flags.Lookup("csv-domain"))
}
