package main

import (
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

var mainCmd = &cobra.Command{
	Use: "bitindex [command]",
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

			defer o.Close()

			w = o
		}

		if err := bitindex.Dump(w, idx); err != nil {
			cmd.Println("Error dumping index:", err)
			os.Exit(1)
		}
	},
}

var statsCmd = &cobra.Command{
	Use: "stats <path>",

	Short: "Outputs the stats of a bitindex.",

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Println("A path to the index file is required.")
			os.Exit(1)
		}

		f, err := os.Open(args[0])

		if err != nil {
			cmd.Println("Error opening file:", err)
			os.Exit(1)
		}

		if err != nil {
			cmd.Println("Error decompressing file:", err)
			os.Exit(1)
		}

		idx := bitindex.NewIndex(nil)

		if err := bitindex.Load(f, idx); err != nil {
			cmd.Println("Error loading index file:", err)
			os.Exit(1)
		}

		cmd.Println("Statistics")
		cmd.Println("* Domain size:", idx.Domain.Size())
		cmd.Println("* Table size:", idx.Table.Size())
		cmd.Println("* Sparsity:", idx.Sparsity()*100)
	},
}

func parseOpFlag(s string) ([]uint32, error) {
	if s == "" {
		return nil, nil
	}

	var (
		n   uint64
		err error
	)

	toks := strings.Split(s, ",")
	ints := make([]uint32, len(toks))

	for i, t := range toks {
		if n, err = strconv.ParseUint(t, 10, 32); err != nil {
			return nil, err
		}

		ints[i] = uint32(n)
	}

	return ints, nil
}

type uint32Set map[uint32]struct{}

func (s uint32Set) Add(is ...uint32) {
	for _, i := range is {
		s[i] = struct{}{}
	}
}

func (s uint32Set) Remove(is ...uint32) {
	for _, i := range is {
		delete(s, i)
	}
}

func (s uint32Set) Clear() {
	for k, _ := range s {
		delete(s, k)
	}
}

func (s uint32Set) Intersect(b uint32Set) uint32Set {
	o := make(uint32Set)

	// Pick the smallest set.
	x := s

	if len(b) < len(s) {
		x = b
	}

	for k, _ := range x {
		if _, ok := b[k]; ok {
			o[k] = struct{}{}
		}
	}

	return o
}

var queryCmd = &cobra.Command{
	Use: "query <path>",

	Short: "Queries an index.",

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Println("A path to the index file is required.")
			os.Exit(1)
		}

		var (
			any, all, nany, nall []uint32
			err                  error
		)

		// Parse operation flags.
		if any, err = parseOpFlag(viper.GetString("query.any")); err != nil {
			cmd.Println("Error parsing --any flag.")
			os.Exit(1)
		}

		if all, err = parseOpFlag(viper.GetString("query.all")); err != nil {
			cmd.Println("Error parsing --all flag.")
			os.Exit(1)
		}

		if nany, err = parseOpFlag(viper.GetString("query.nany")); err != nil {
			cmd.Println("Error parsing --nany flag.")
			os.Exit(1)
		}

		if nall, err = parseOpFlag(viper.GetString("query.nall")); err != nil {
			cmd.Println("Error parsing --nall flag.")
			os.Exit(1)
		}

		if len(any) == 0 && len(all) == 0 && len(nany) == 0 && len(nall) == 0 {
			cmd.Println("At least one operation must be specified.")
			os.Exit(1)
		}

		f, err := os.Open(args[0])

		if err != nil {
			cmd.Println("Error opening file:", err)
			os.Exit(1)
		}

		idx := bitindex.NewIndex(nil)

		if err := bitindex.Load(f, idx); err != nil {
			cmd.Println("Error loading index file:", err)
			os.Exit(1)
		}

		var (
			set  uint32Set
			tmp  = make(uint32Set)
			keys []uint32
		)

		if any != nil {
			if keys, err = idx.Any(any...); err != nil {
				cmd.Printf("Operation failed (any): %s\n", err)
				os.Exit(1)
			}

			if set == nil {
				set = make(uint32Set, len(keys))
				set.Add(keys...)
			} else {
				tmp.Add(keys...)
				set = set.Intersect(tmp)
				tmp.Clear()
			}
		}

		if all != nil {
			if keys, err = idx.All(all...); err != nil {
				cmd.Printf("Operation failed (all): %s\n", err)
				os.Exit(1)
			}

			if set == nil {
				set = make(uint32Set, len(keys))
				set.Add(keys...)
			} else {
				tmp.Add(keys...)
				set = set.Intersect(tmp)
				tmp.Clear()
			}
		}

		if nany != nil {
			if keys, err = idx.NotAny(nany...); err != nil {
				cmd.Printf("Operation failed (nany): %s\n", err)
				os.Exit(1)
			}

			if set == nil {
				set = make(uint32Set, len(keys))
				set.Add(keys...)
			} else {
				tmp.Add(keys...)
				set = set.Intersect(tmp)
				tmp.Clear()
			}
		}

		if nall != nil {
			if keys, err = idx.NotAll(nall...); err != nil {
				cmd.Printf("Operation failed (nall): %s\n", err)
				os.Exit(1)
			}

			if set == nil {
				set = make(uint32Set, len(keys))
				set.Add(keys...)
			} else {
				tmp.Add(keys...)
				set = set.Intersect(tmp)
				tmp.Clear()
			}
		}

		for k, _ := range set {
			fmt.Println(k)
		}
	},
}

func main() {
	mainCmd.AddCommand(buildCmd)
	mainCmd.AddCommand(statsCmd)
	mainCmd.AddCommand(queryCmd)

	mainCmd.Execute()
}

func init() {
	initIndexFlags()
	initQueryFlags()
}

func initIndexFlags() {
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

func initQueryFlags() {
	flags := queryCmd.Flags()

	flags.String("any", "", "Applies the any operation.")
	flags.String("all", "", "Applies the all operation.")
	flags.String("nany", "", "Applies the not any operation.")
	flags.String("nall", "", "Applies the not all operation.")

	viper.BindPFlag("query.any", flags.Lookup("any"))
	viper.BindPFlag("query.all", flags.Lookup("all"))
	viper.BindPFlag("query.nany", flags.Lookup("nany"))
	viper.BindPFlag("query.nall", flags.Lookup("nall"))
}
