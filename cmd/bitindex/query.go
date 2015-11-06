package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chop-dbhi/bitindex"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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

var queryCmd = &cobra.Command{
	Use: "query <index>",

	Short: "Queries an index.",

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Println("An index file is required.")
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

		var idx *bitindex.Index

		if idx, err = bitindex.LoadIndex(f); err != nil {
			cmd.Println("Error loading index file:", err)
			os.Exit(1)
		}

		// Query time.
		t0 := time.Now()

		res, err := idx.Query(any, all, nany, nall)

		if err != nil {
			cmd.Println("Error with query:", err)
			os.Exit(1)
		}

		cmd.Printf("Time: %s\n", time.Now().Sub(t0))

		for _, k := range res {
			fmt.Println(k)
		}
	},
}

func init() {
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
