package main

import "github.com/spf13/cobra"

var mainCmd = &cobra.Command{
	Use: "bitindex [command]",
}

func main() {
	mainCmd.AddCommand(buildCmd)
	mainCmd.AddCommand(domainCmd)
	mainCmd.AddCommand(keysCmd)
	mainCmd.AddCommand(statsCmd)
	mainCmd.AddCommand(queryCmd)
	mainCmd.AddCommand(httpCmd)

	mainCmd.Execute()
}
