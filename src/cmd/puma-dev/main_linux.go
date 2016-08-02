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
	fDebug    = flag.Bool("debug", false, "enable debug output")
	fDomains  = flag.String("d", "dev", "domains to handle, separate with :")
	fHTTPPort = flag.Int("http-port", 9280, "port to listen on http for")
	fTLSPort  = flag.Int("https-port", 9283, "port to listen on https for")
	fSysBind  = flag.Bool("sysbind", false, "bind to ports 80 and 443")
	fDir      = flag.String("dir", "~/.puma-dev", "directory to watch for apps")
	fTimeout  = flag.Duration("timeout", 15*60*time.Second, "how long to let an app idle for")
)

func main() {
	flag.Parse()

	domains := strings.Split(*fDomains, ":")

	if *fSysBind {
		*fHTTPPort = 80
		*fTLSPort = 443
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

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)

	go func() {
		<-stop
		fmt.Printf("! Shutdown requested\n")
		pool.Purge()
		os.Exit(0)
	}()

	err = dev.SetupOurCert()
	if err != nil {
		log.Fatalf("Unable to setup TLS cert: %s", err)
	}

	fmt.Printf("* Directory for apps: %s\n", dir)
	fmt.Printf("* Domains: %s\n", strings.Join(domains, ", "))
	fmt.Printf("* HTTP Server port: %d\n", *fHTTPPort)
	fmt.Printf("* HTTPS Server port: %d\n", *fTLSPort)

	var http dev.HTTPServer

	http.Address = fmt.Sprintf("127.0.0.1:%d", *fHTTPPort)
	http.TLSAddress = fmt.Sprintf("127.0.0.1:%d", *fTLSPort)
	http.Pool = &pool
	http.Debug = *fDebug

	http.Setup()

	fmt.Printf("! Puma dev listening on http and https\n")

	go http.ServeTLS()

	err = http.Serve()
	if err != nil {
		log.Fatalf("Error listening: %s", err)
	}
}
