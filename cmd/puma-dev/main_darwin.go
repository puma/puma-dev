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

	"github.com/puma/puma-dev/dev"
	"github.com/puma/puma-dev/homedir"
)

var (
	fDebug    = flag.Bool("debug", false, "enable debug output")
	fDomains  = flag.String("d", "test", "domains to handle, separate with :, defaults to test")
	fDNSPort  = flag.Int("dns-port", 9253, "port to listen on dns for")
	fHTTPPort = flag.Int("http-port", 9280, "port to listen on http for")
	fTLSPort  = flag.Int("https-port", 9283, "port to listen on https for")
	fDir      = flag.String("dir", "~/.puma-dev", "directory to watch for apps")
	fTimeout  = flag.Duration("timeout", 15*60*time.Second, "how long to let an app idle for")
	fPow      = flag.Bool("pow", false, "Mimic pow's settings")
	fLaunch   = flag.Bool("launchd", false, "Use socket from launchd")

	fSetup = flag.Bool("setup", false, "Run system setup")
	fStop  = flag.Bool("stop", false, "Stop all puma-dev servers")

	fInstall     = flag.Bool("install", false, "Install puma-dev as a user service")
	fInstallPort = flag.Int("install-port", 80, "Port to run puma-dev on when installed")
	fInstallTLS  = flag.Int("install-https-port", 443, "Port to run puma-dev for SSL on when installed")

	fCleanup   = flag.Bool("cleanup", false, "Cleanup old system settings")
	fUninstall = flag.Bool("uninstall", false, "Uninstall puma-dev as a user service")
)

const (
	LaunchAgentDirPath = "~/Library/LaunchAgents"
	LogFilePath        = "~/Library/Logs/puma-dev.log"
)

func main() {
	flag.Parse()

	allCheck()

	if *fPow {
		*fDomains = "dev"
		*fDir = "~/.pow"
	}

	domains := strings.Split(*fDomains, ":")

	if *fCleanup {
		dev.Cleanup()
		return
	}

	if *fUninstall {
		dev.Uninstall(LaunchAgentDirPath, domains)
		return
	}

	if *fInstall {
		err := dev.InstallIntoSystem(&dev.InstallIntoSystemArgs{
			ApplinkDirPath:     *fDir,
			Domains:            *fDomains,
			LaunchAgentDirPath: LaunchAgentDirPath,
			ListenPort:         *fInstallPort,
			LogfilePath:        LogFilePath,
			Timeout:            (*fTimeout).String(),
			TlsPort:            *fInstallTLS,
		})

		if err != nil {
			log.Fatalf("Unable to install into system: %s", err)
		}
		return
	}

	if *fSetup {
		err := dev.Setup()
		if err != nil {
			log.Fatalf("Unable to configure OS X resolver: %s", err)
		}
		return
	}

	if *fStop {
		err := dev.Stop()
		if err != nil {
			log.Fatalf("Unable to stop puma-dev servers: %s", err)
		}
		return
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

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)

	go func() {
		<-shutdown
		fmt.Printf("! Shutdown requested\n")
		pool.Purge()
		os.Exit(0)
	}()

	err = dev.ConfigureResolver(domains, *fDNSPort)
	if err != nil {
		log.Fatalf("Unable to configure OS X resolver: %s", err)
	}

	err = dev.SetupOurCert()
	if err != nil {
		log.Fatalf("Unable to setup TLS cert: %s", err)
	}

	fmt.Printf("* Directory for apps: %s\n", dir)
	fmt.Printf("* Domains: %s\n", strings.Join(domains, ", "))
	fmt.Printf("* DNS Server port: %d\n", *fDNSPort)

	if *fLaunch {
		fmt.Printf("* HTTP Server port: inherited from launchd\n")
		fmt.Printf("* HTTPS Server port: inherited from launchd\n")
	} else {
		fmt.Printf("* HTTP Server port: %d\n", *fHTTPPort)
		fmt.Printf("* HTTPS Server port: %d\n", *fTLSPort)
	}

	dns := dev.NewDNSResponder(fmt.Sprintf("127.0.0.1:%d", *fDNSPort), domains)
	go func() {
		if err := dns.Serve(); err != nil {
			fmt.Printf("! DNS Server failed: %v\n", err)
		}
	}()

	var http dev.HTTPServer

	http.Address = fmt.Sprintf("127.0.0.1:%d", *fHTTPPort)
	http.TLSAddress = fmt.Sprintf("127.0.0.1:%d", *fTLSPort)
	http.Pool = &pool
	http.Debug = *fDebug
	http.Events = &events

	http.Setup()

	var (
		socketName    string
		tlsSocketName string
	)

	if *fLaunch {
		socketName = "Socket"
		tlsSocketName = "SocketTLS"
	}

	fmt.Printf("! Puma dev running...\n")

	go func() {
		if err := http.ServeTLS(tlsSocketName); err != nil {
			fmt.Printf("! HTTPS Server failed: %v\n", err)
		}
	}()

	err = http.Serve(socketName)
	if err != nil {
		log.Fatalf("! HTTP Server failed: %s", err)
	}
}
