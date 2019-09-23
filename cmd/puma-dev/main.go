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
	if *fVersion {
		fmt.Printf("Version: %s (%s)\n", Version, runtime.Version())
		os.Exit(0)
	}

	if flag.NArg() > 0 {
		err := command()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}

		os.Exit(0)
		return
	}
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()

		fmt.Fprintf(os.Stderr, "\nAvailable subcommands: link\n")
	}
}
