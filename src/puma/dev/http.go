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

	transport *httpu.Transport
}

func (h *HTTPServer) Setup() {
	h.transport = &httpu.Transport{
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

func pruneSub(name string) string {
	dot := strings.IndexByte(name, '.')
	if dot == -1 {
		return ""
	}

	return name[dot+1:]
}

func (h *HTTPServer) hostForApp(name string) (string, string, error) {
	for name != "" {
		app, err := h.Pool.App(name)
		if err != nil {
			if err == ErrUnknownApp {
				name = pruneSub(name)
				continue
			}

			panic(err)
		}

		return app.Scheme, app.Address(), nil
	}

	app, err := h.Pool.App("default")
	if err != nil {
		return "", "", err
	}

	return app.Scheme, app.Address(), nil
}

func (h *HTTPServer) director(req *http.Request) error {
	dot := strings.LastIndexByte(req.Host, '.')
	name := req.Host[:dot]

	var err error
	req.URL.Scheme, req.URL.Host, err = h.hostForApp(name)
	return err
}
