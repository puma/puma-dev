package dev

import (
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

type HTTPServer struct {
	Address string
	Pool    *AppPool

	proxy *httputil.ReverseProxy
}

func (h *HTTPServer) hostForApp(name string) string {
	app, err := h.Pool.App(name)
	if err != nil {
		panic(err)
	}

	return app.Address()
}

func (h *HTTPServer) director(req *http.Request) {
	parts := strings.Split(req.Host, ".")

	app := parts[len(parts)-2]

	req.URL.Scheme = "http"
	req.URL.Host = h.hostForApp(app)
}

func (h *HTTPServer) Serve() error {
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	h.proxy = &httputil.ReverseProxy{
		Director:  h.director,
		Transport: transport,
	}

	h.proxy.FlushInterval = 1 * time.Second
	h.proxy.Transport = transport

	serv := http.Server{
		Addr:    h.Address,
		Handler: h.proxy,
	}

	return serv.ListenAndServe()
}
