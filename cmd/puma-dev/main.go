package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
)

var (
	EarlyExitClean = CommandResult{0, true}
	EarlyExitError = CommandResult{1, true}
	Continue       = CommandResult{-1, false}

	fVersion = flag.Bool("V", false, "display version info")
	Version  = "devel"
)

type CommandResult struct {
	exitStatusCode int
	shouldExit     bool
}

type ByDecreasingTLDComplexity []string

func (a ByDecreasingTLDComplexity) Len() int      { return len(a) }
func (a ByDecreasingTLDComplexity) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByDecreasingTLDComplexity) Less(i, j int) bool {
	return strings.Count(a[i], ".") > strings.Count(a[j], ".")
}

func allCheck() {
	if result := execWithExitStatus(); result.shouldExit {
		os.Exit(result.exitStatusCode)
	}
}

func execWithExitStatus() CommandResult {
	if *fVersion {
		fmt.Printf("Version: %s (%s)\n", Version, runtime.Version())
		return EarlyExitClean
	}

	if flag.NArg() > 0 {
		err := command()

		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return EarlyExitError
		}

		return EarlyExitClean
	}

	return Continue
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()

		fmt.Fprintf(os.Stderr, "\nAvailable subcommands: link\n")
	}
}
