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
	if status := execWithStatus(); status >= 0 {
		os.Exit(0)
	}
}

func execWithStatus() int {
	if *fVersion {
		fmt.Printf("Version: %s (%s)\n", Version, runtime.Version())
		return 0
	}

	if flag.NArg() > 0 {
		err := command()

		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return 1
		}

		return 0
	}

	return -1
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()

		fmt.Fprintf(os.Stderr, "\nAvailable subcommands: link\n")
	}
}
