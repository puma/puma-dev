package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/puma/puma-dev/homedir"
	"github.com/vektra/errors"
)

func command() error {
	switch flag.Arg(0) {
	case "link":
		return link()
	default:
		return fmt.Errorf("Unknown command: %s\n", flag.Arg(0))
	}
}

func link() error {
	fs := flag.NewFlagSet("link", flag.ExitOnError)
	name := fs.String("n", "", "name to link app as")

	err := fs.Parse(flag.Args()[1:])
	if err != nil {
		return err
	}

	var dir string

	if fs.NArg() == 0 {
		dir, err = os.Getwd()
		if err != nil {
			return err
		}

	} else {
		dir, err = filepath.Abs(fs.Arg(0))
		if err != nil {
			return err
		}

		stat, err := os.Stat(dir)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("Invalid directory: %s", dir)
			}

			return err
		}

		if !stat.IsDir() {
			return fmt.Errorf("Invalid directory: %s", dir)
		}
	}

	if *name == "" {
		*name = strings.Replace(filepath.Base(dir), ".", "_", -1)
	}

	dest, err := homedir.Expand(filepath.Join(*fDir, *name))
	if err != nil {
		return err
	}

	_, err = os.Stat(dest)
	if err == nil {
		dest, err := os.Readlink(dest)
		if err == nil {
			fmt.Printf("! App '%s' already exists, pointed at '%s'\n", *name, dest)
		} else {
			fmt.Printf("! App '%s' already exists'\n", *name)
		}
		return nil
	}

	err = os.Symlink(dir, dest)
	if err != nil {
		return errors.Context(err, "creating symlink")
	}

	fmt.Printf("+ App '%s' created, linked to '%s'\n", *name, dir)

	return nil
}
