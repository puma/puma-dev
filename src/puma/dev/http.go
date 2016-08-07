package dev

import (
	"net"
	"net/http"
	"puma/httpu"
	"strings"
	"time"
)

type HTTPServer struct {
	Address    string
	TLSAddress string
	Pool       *AppPool
	Debug      bool

	transport *httpu.Transport
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

func (h *HTTPServer) director(req *http.Request) error {
	dot := strings.LastIndexByte(req.Host, '.')

	var name string
	if dot == -1 {
		name = req.Host
	} else {
		name = req.Host[:dot]
	}

	var err error
	req.URL.Scheme, req.URL.Host, err = h.hostForApp(name)
	return err
}
