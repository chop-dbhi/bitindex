package main

import (
	"fmt"
	"os"

	"github.com/chop-dbhi/bitindex"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var keysCmd = &cobra.Command{
	Use: "keys <index>",

	Short: "Outputs information about the keys.",

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
		cmd.Println("* Length:", idx.Table.Size())
		cmd.Println("* Bytes:", idx.Table.Bytes())

		if viper.GetBool("keys.keys") {
			for _, k := range idx.Table.Keys() {
				fmt.Fprintln(os.Stdout, k)
			}
		}
	},
}

func init() {
	flags := keysCmd.Flags()

	flags.Bool("keys", false, "Ouputs the keys to stdout.")

	viper.BindPFlag("keys.keys", flags.Lookup("keys"))
}
