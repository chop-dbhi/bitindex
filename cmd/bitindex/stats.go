package main

import (
	"os"

	"github.com/chop-dbhi/bitindex"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use: "stats <index>",

	Short: "Outputs the stats of a bitindex.",

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Println("An index file is required.")
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

		cmd.Println("Statistics")
		cmd.Println("* Domain size:", idx.Domain.Size())
		cmd.Println("* Table size:", idx.Table.Size())
		cmd.Println("* Sparsity:", idx.Sparsity()*100)
	},
}
