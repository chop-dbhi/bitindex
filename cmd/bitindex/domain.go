package main

import (
	"fmt"
	"os"

	"github.com/chop-dbhi/bitindex"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var domainCmd = &cobra.Command{
	Use: "domain <index>",

	Short: "Outputs information about the domain.",

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

		var d *bitindex.Domain

		if d, err = bitindex.LoadDomain(f); err != nil {
			cmd.Println("Error loading index file:", err)
			os.Exit(1)
		}

		cmd.Println("Statistics")
		cmd.Println("* Length:", d.Size())
		cmd.Println("* Bytes:", d.Bytes())

		if viper.GetBool("domain.members") {
			for _, m := range d.Members() {
				fmt.Fprintln(os.Stdout, m)
			}
		}
	},
}

func init() {
	flags := domainCmd.Flags()

	flags.Bool("members", false, "Ouputs the members to stdout.")

	viper.BindPFlag("domain.members", flags.Lookup("members"))
}
