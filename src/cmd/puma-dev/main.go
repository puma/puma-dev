package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/mitchellh/go-homedir"

	"puma/dev"
)

var (
	fDomains  = flag.String("d", "pdev", "domains to handle, separate with :")
	fPort     = flag.Int("dns-port", 9253, "port to listen on dns for")
	fHTTPPort = flag.Int("http-port", 9280, "port to listen on http for")
	fDir      = flag.String("dir", "~/.puma-dev", "directory to watch for apps")
	fTimeout  = flag.Duration("timeout", 15*60*time.Second, "how long to let an app idle for")
	fPow      = flag.Bool("pow", false, "Mimic pow's settings")

	fSetup         = flag.Bool("setup", false, "Run system setup")
	fSetupSkipHTTP = flag.Bool("setup-skip-80", false, "Indicate if a firewall rule to redirect port 80 to our port should be skipped")

	fInstall = flag.Bool("install", false, "Install puma-dev as a system service")
)

func main() {
	flag.Parse()

	if *fInstall {
		err := dev.InstallIntoSystem(*fSetupSkipHTTP)
		if err != nil {
			log.Fatalf("Unable to install into system: %s", err)
		}
		return
	}

	if *fSetup {
		err := dev.Setup(*fSetupSkipHTTP)
		if err != nil {
			log.Fatalf("Unable to configure OS X resolver: %s", err)
		}
		return
	}

	if *fPow {
		*fDomains = "dev"
		*fDir = "~/.pow"
	}

	dir, err := homedir.Expand(*fDir)
	if err != nil {
		log.Fatalf("Unable to expand dir: %s", err)
	}

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatalf("Unable to create dir '%s': %s", dir, err)
	}

	var pool dev.AppPool
	pool.Dir = dir
	pool.IdleTime = *fTimeout

	purge := make(chan os.Signal, 1)

	signal.Notify(purge, syscall.SIGUSR1)

	go func() {
		for {
			<-purge
			pool.Purge()
		}
	}()

	domains := strings.Split(*fDomains, ":")

	err = dev.ConfigureResolver(domains, *fPort)
	if err != nil {
		log.Fatalf("Unable to configure OS X resolver: %s", err)
	}

	fmt.Printf("* Directory for apps: %s\n", dir)
	fmt.Printf("* Domains: %s\n", strings.Join(domains, ", "))
	fmt.Printf("* DNS Server port: %d\n", *fPort)
	fmt.Printf("* HTTP Server port: %d\n", *fHTTPPort)

	var dns dev.DNSResponder

	dns.Address = fmt.Sprintf("127.0.0.1:%d", *fPort)

	go dns.Serve(domains)

	var http dev.HTTPServer

	http.Address = fmt.Sprintf("127.0.0.1:%d", *fHTTPPort)
	http.Pool = &pool

	fmt.Printf("! Puma dev listening\n")
	http.Serve()
}
