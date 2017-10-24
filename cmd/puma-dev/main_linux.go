package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/puma/puma-dev/dev"
	"github.com/puma/puma-dev/homedir"
)

var (
	fDebug    = flag.Bool("debug", false, "enable debug output")
	fDomains  = flag.String("d", DefaultDomains, "domains to handle, separate with :")
	fHTTPPort = flag.Int("http-port", DefaultHttpPort, "port to listen on http for")
	fTLSPort  = flag.Int("https-port", DefaultTlsPort, "port to listen on https for")
	fSysBind  = flag.Bool("sysbind", false, "bind to ports 80 and 443")
	fDir      = flag.String("dir", DefaultDir, "directory to watch for apps")
	fTimeout  = flag.Duration("timeout", 15*60*time.Second, "how long to let an app idle for")
	fStop     = flag.Bool("stop", false, "Stop all puma-dev servers")
)

func main() {
	flag.Parse()

	parseEnvFlags()

	allCheck()

	domains := strings.Split(*fDomains, ":")

	if *fStop {
		err := dev.Stop()
		if err != nil {
			log.Fatalf("Unable to stop puma-dev servers: %s", err)
		}
		return
	}

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

	var events dev.Events

	var pool dev.AppPool
	pool.Dir = dir
	pool.IdleTime = *fTimeout
	pool.Events = &events

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

	http.Address = fmt.Sprintf(":%d", *fHTTPPort)
	http.TLSAddress = fmt.Sprintf(":%d", *fTLSPort)
	http.Pool = &pool
	http.Debug = *fDebug
	http.Events = &events

	http.Setup()

	fmt.Printf("! Puma dev listening on http and https\n")

	go http.ServeTLS()

	err = http.Serve()
	if err != nil {
		log.Fatalf("Error listening: %s", err)
	}
}

func parseEnvFlags() {
	if !*fDebug {
		envVal := os.Getenv("PUMA_DEV_DEBUG")
		if envVal != "" {
			*fDebug = true
		}
	}

	if *fDomains == DefaultDomains {
		envVal := os.Getenv("PUMA_DEV_DOMAINS")

		if envVal != "" {
			*fDomains = envVal
		}
	}

	if *fHTTPPort == DefaultHttpPort {
		envVal := os.Getenv("PUMA_DEV_HTTP_PORT")

		if envVal != "" {
			*fHTTPPort, _ = strconv.Atoi(envVal)
		}
	}

	if *fTLSPort == DefaultTlsPort {
		envVal := os.Getenv("PUMA_DEV_HTTPS_PORT")

		if envVal != "" {
			*fTLSPort, _ = strconv.Atoi(envVal)
		}
	}

	if *fDir == DefaultDir {
		envVal := os.Getenv("PUMA_DEV_DIR")

		if envVal != "" {
			*fDir = envVal
		}
	}

	if *fTimeout == 15*60*time.Second {
		envVal := os.Getenv("PUMA_DEV_TIMEOUT")

		if envVal != "" {
			parsedVal, _ := strconv.Atoi(envVal)
			*fTimeout = time.Duration(parsedVal) * time.Second
		}
	}
}
