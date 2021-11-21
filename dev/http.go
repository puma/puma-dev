package dev

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bmizerany/pat"
	"github.com/puma/puma-dev/httpu"
	"github.com/puma/puma-dev/httputil"
)

type HTTPServer struct {
	Address            string
	TLSAddress         string
	Pool               *AppPool
	Debug              bool
	Events             *Events
	IgnoredStaticPaths []string
	Domains            []string

	mux       *pat.PatternServeMux
	transport *httpu.Transport
	proxy     *httputil.ReverseProxy
}

func (h *HTTPServer) Setup() {
	h.transport = &httpu.Transport{
		Dial: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 10 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	h.Pool.AppClosed = h.AppClosed

	h.proxy = &httputil.ReverseProxy{
		Proxy:         h.proxyReq,
		Transport:     h.transport,
		FlushInterval: 1 * time.Second,
		Debug:         h.Debug,
	}

	h.mux = pat.New()

	h.mux.Get("/status", http.HandlerFunc(h.status))
	h.mux.Get("/events", http.HandlerFunc(h.events))
}

func (h *HTTPServer) AppClosed(app *App) {
	// Whenever an app is closed, wipe out all idle conns. This
	// obviously closes down more than just this one apps connections
	// but that's ok.
	h.transport.CloseIdleConnections()
}

type ByDecreasingTLDComplexity []string

func (a ByDecreasingTLDComplexity) Len() int      { return len(a) }
func (a ByDecreasingTLDComplexity) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByDecreasingTLDComplexity) Less(i, j int) bool {
	return strings.Count(a[i], ".") > strings.Count(a[j], ".")
}

func (h *HTTPServer) removeTLD(host string) string {
	colon := strings.LastIndexByte(host, ':')
	if colon != -1 {
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
	}

	if strings.HasSuffix(host, ".xip.io") || strings.HasSuffix(host, ".nip.io") {
		parts := strings.Split(host, ".")
		if len(parts) < 6 {
			return ""
		}

		name := strings.Join(parts[:len(parts)-6], ".")

		return name
	}

	if !sort.IsSorted(ByDecreasingTLDComplexity(h.Domains)) {
		sort.Sort(ByDecreasingTLDComplexity(h.Domains))
	}

	for _, tld := range h.Domains {
		if strings.HasSuffix(host, "."+tld) {
			return strings.TrimSuffix(host, "."+tld)
		}
	}

	dot := strings.LastIndexByte(host, '.')

	if dot == -1 {
		return host
	} else {
		return host[:dot]
	}
}

func (h *HTTPServer) proxyReq(w http.ResponseWriter, req *http.Request) error {
	name := h.removeTLD(req.Host)

	app, err := h.Pool.FindAppByDomainName(name)
	if err != nil {
		if err == ErrUnknownApp {
			h.Events.Add("unknown_app", "name", name, "host", req.Host)
		} else {
			h.Events.Add("lookup_error", "error", err.Error())
		}

		return err
	}

	if h.shouldServePublicPathForApp(app, req) {
		safeURLPath := path.Clean(req.URL.Path)
		path := filepath.Join(app.dir, "public", safeURLPath)

		fi, err := os.Stat(path)
		if err == nil && !fi.IsDir() {
			if ofile, err := os.Open(path); err == nil {
				http.ServeContent(w, req, req.URL.Path, fi.ModTime(), io.ReadSeeker(ofile))
				return httputil.ErrHandled
			}
		}
	}

	req.URL.Scheme, req.URL.Host = app.Scheme, app.Address()
	return err
}

func (h *HTTPServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if h.Debug {
		fmt.Fprintf(os.Stderr, "%s: %s '%s' (host=%s)\n",
			time.Now().Format(time.RFC3339Nano),
			req.Method, req.URL.Path, req.Host)
	}

	if req.Host == "puma-dev" {
		h.mux.ServeHTTP(w, req)
	} else {
		h.proxy.ServeHTTP(w, req)
	}
}

func (h *HTTPServer) shouldServePublicPathForApp(a *App, req *http.Request) bool {
	reqPath := path.Clean(req.URL.Path)

	if !a.Public {
		return false
	}

	if reqPath == "/" {
		return false
	}

	for _, ignoredPath := range h.IgnoredStaticPaths {
		if strings.HasPrefix(reqPath, ignoredPath) {
			if h.Debug {
				fmt.Fprintf(os.Stdout, "Not serving '%s' as it matches a path in no-serve-public-paths\n", reqPath)
			}
			return false
		}
	}

	return true
}

func (h *HTTPServer) status(w http.ResponseWriter, req *http.Request) {
	type appStatus struct {
		Scheme  string `json:"scheme"`
		Address string `json:"address"`
		Status  string `json:"status"`
		Log     string `json:"log"`
	}

	statuses := map[string]appStatus{}

	h.Pool.ForApps(func(a *App) {
		var status string

		switch a.Status() {
		case Dead:
			status = "dead"
		case Booting:
			status = "booting"
		case Running:
			status = "running"
		default:
			status = "unknown"
		}

		statuses[a.Name] = appStatus{
			Scheme:  a.Scheme,
			Address: a.Address(),
			Status:  status,
			Log:     a.Log(),
		}
	})

	json.NewEncoder(w).Encode(statuses)
}

func (h *HTTPServer) events(w http.ResponseWriter, req *http.Request) {
	h.Events.WriteTo(w)
}
