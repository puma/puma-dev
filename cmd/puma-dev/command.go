package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/puma/puma-dev/homedir"
	"github.com/vektra/errors"
)

func command() error {
	switch flag.Arg(0) {
	case "link":
		return link()
	case "status":
		return status()
	default:
		return fmt.Errorf("Unknown command: %s\n", flag.Arg(0))
	}
}

// App is a running application.
type App struct {
	Status  string `json:"status"`
	Scheme  string `json:"scheme"`
	Address string `json:"address"`
}

func status() error {
	var port string

	if *fHTTPPort != 9280 {
		port = fmt.Sprintf(":%d", *fHTTPPort)
	}

	client := &http.Client{}
	url := fmt.Sprintf("http://localhost%s/status", port)
	req, err := http.NewRequest("GET", url, nil)
	req.Host = "puma-dev"
	w := tabwriter.NewWriter(os.Stdout, 20, 4, 1, ' ', 0)

	if err != nil {
		return err
	}

	res, err := client.Do(req)

	if err != nil {
		fmt.Printf("Unable to lookup puma-dev status. Is puma-dev listening on port %d?\n", *fHTTPPort)
		return nil
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return err
	}

	var apps map[string]*App
	err = json.Unmarshal(body, &apps)

	if err != nil {
		return err
	}

	if len(apps) > 0 {

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", "NAME", "STATUS", "ADDRESS", "SCHEME")

		for name, app := range apps {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", name, app.Status, app.Address, app.Scheme)
		}

		w.Flush()
	} else {
		fmt.Println("No apps are currently running.")
	}

	return nil
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
