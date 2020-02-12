package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
)

var fVersion = flag.Bool("V", false, "display version info")
var Version = "devel"

func allCheck() {
	if status, shouldExit := execWithExitStatus(); shouldExit {
		os.Exit(status)
	}
}

func execWithExitStatus() (int, bool) {
	if *fVersion {
		fmt.Printf("Version: %s (%s)\n", Version, runtime.Version())
		return 0, true
	}

	if flag.NArg() > 0 {
		err := command()

		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return 1, true
		}

		return 0, true
	}

	return -1, false
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()

		fmt.Fprintf(os.Stderr, "\nAvailable subcommands: link\n")
	}
}
