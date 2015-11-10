package main

import (
	"os"

	"github.com/davecheney/profile"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var mainCmd = &cobra.Command{
	Use: "bitindex [command]",
}

func init() {
	flags := mainCmd.PersistentFlags()

	flags.Bool("prof", false, "Enable profiling.")
	flags.String("prof-path", "./prof", "The path to store output profiles.")
	flags.Float32("smallest-threshold", 0.6, "The ratio of the set size for the complement to be returned.")

	viper.BindPFlag("main.prof", flags.Lookup("prof"))
	viper.BindPFlag("main.prof-path", flags.Lookup("prof-path"))
	viper.BindPFlag("main.smallest-threshold", flags.Lookup("smallest-threshold"))
}

func main() {
	mainCmd.AddCommand(versionCmd)
	mainCmd.AddCommand(buildCmd)
	mainCmd.AddCommand(domainCmd)
	mainCmd.AddCommand(keysCmd)
	mainCmd.AddCommand(statsCmd)
	mainCmd.AddCommand(queryCmd)
	mainCmd.AddCommand(httpCmd)

	// Parse flags early so we can start the profiler.
	mainCmd.ParseFlags(os.Args)

	if viper.GetBool("main.prof") {
		defer profile.Start(&profile.Config{
			Quiet:       true,
			CPUProfile:  true,
			MemProfile:  true,
			ProfilePath: viper.GetString("main.prof-path"),
		}).Stop()
	}

	mainCmd.Execute()
}
