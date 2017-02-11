package dev

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/bmizerany/pat"
	"github.com/puma/puma-dev/httpu"
	"github.com/puma/puma-dev/httputil"
)

type HTTPServer struct {
	Address    string
	TLSAddress string
	Pool       *AppPool
	Debug      bool
	Events     *Events

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

func pruneSub(name string) string {
	dot := strings.IndexByte(name, '.')
	if dot == -1 {
		return ""
	}

	return name[dot+1:]
}

func (h *HTTPServer) findApp(name string) (*App, error) {
	var (
		app *App
		err error
	)

	for name != "" {
		app, err = h.Pool.App(name)
		if err != nil {
			if err == ErrUnknownApp {
				name = pruneSub(name)
				continue
			}

			return nil, err
		}

		break
	}

	if app == nil {
		app, err = h.Pool.App("default")
		if err != nil {
			return nil, err
		}
	}

	err = app.WaitTilReady()
	if err != nil {
		return nil, err
	}

	return app, nil
}

func (h *HTTPServer) hostForApp(name string) (string, string, error) {
	var (
		app *App
		err error
	)

	for name != "" {
		app, err = h.Pool.App(name)
		if err != nil {
			if err == ErrUnknownApp {
				name = pruneSub(name)
				continue
			}

			return "", "", err
		}

		break
	}

	if app == nil {
		app, err = h.Pool.App("default")
		if err != nil {
			return "", "", err
		}
	}

	err = app.WaitTilReady()
	if err != nil {
		return "", "", err
	}

	return app.Scheme, app.Address(), nil
}

func (h *HTTPServer) removeTLD(host string) string {
	colon := strings.LastIndexByte(host, ':')
	if colon != -1 {
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
	}

	if strings.HasSuffix(host, ".xip.io") {
		parts := strings.Split(host, ".")
		if len(parts) < 6 {
			return ""
		}

		name := strings.Join(parts[:len(parts)-6], ".")

		return name
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

	app, err := h.findApp(name)
	if err != nil {
		if err == ErrUnknownApp {
			h.Events.Add("unknown_app", "name", name, "host", req.Host)
		} else {
			h.Events.Add("lookup_error", "error", err.Error())
		}

		return err
	}

	if app.Public && req.URL.Path != "/" {
		path := filepath.Join(app.dir, "public", path.Clean(req.URL.Path))

		fi, err := os.Stat(path)
		if err == nil && !fi.IsDir() {
			http.ServeFile(w, req, path)
			return httputil.ErrHandled
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
