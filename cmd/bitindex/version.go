package main

import (
	"fmt"

	"github.com/blang/semver"
	"github.com/spf13/cobra"
)

// SemVer components.
const (
	progMajor        = 0
	progMinor        = 1
	progPatch        = 0
	progReleaseLevel = "beta"
	progReleaseNum   = 1
)

var (
	// Populated at build time. See the Makefile for details.
	// Note, in environments where the git information is not
	// available, these will not be populated.
	progBuild string

	// Full semantic version for the service.
	progVersion = semver.Version{
		Major: progMajor,
		Minor: progMinor,
		Patch: progPatch,
		Pre: []semver.PRVersion{{
			VersionStr: progReleaseLevel,
		}, {
			VersionNum: progReleaseNum,
			IsNum:      true,
		}},
	}
)

func init() {
	// Add the build if available.
	if progBuild != "" {
		progVersion.Build = []string{progBuild}
	}
}

var versionCmd = &cobra.Command{
	Use: "version",

	Short: "Prints the version.",

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(progVersion)
	},
}
