package dev

import (
	"crypto/tls"
	"net"
	"net/http"
	"puma/dev/launch"
	"puma/httpu"
	"puma/httputil"
	"strings"
	"time"

	"gopkg.in/tomb.v2"
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

func (h *HTTPServer) ServeTLS(launchdSocket string) error {
	proxy := &httputil.ReverseProxy{
		ForwardProto:  "https",
		Director:      h.director,
		Transport:     h.transport,
		FlushInterval: 1 * time.Second,
	}

	certCache := NewCertCache()

	tlsConfig := &tls.Config{
		GetCertificate: certCache.GetCertificate,
	}

	serv := http.Server{
		Addr:      h.TLSAddress,
		Handler:   proxy,
		TLSConfig: tlsConfig,
	}

	if launchdSocket == "" {
		return serv.ListenAndServeTLS("", "")
	}

	listeners, err := launch.SocketListeners(launchdSocket)
	if err != nil {
		return err
	}

	var t tomb.Tomb

	for i, l := range listeners {
		tl := tls.NewListener(l, tlsConfig)
		listeners[i] = tl
	}

	for _, l := range listeners {
		t.Go(func() error {
			return serv.Serve(l)
		})
	}

	return t.Wait()

}

func (h *HTTPServer) Serve(launchdSocket string) error {
	proxy := &httputil.ReverseProxy{
		Director:      h.director,
		Transport:     h.transport,
		FlushInterval: 1 * time.Second,
	}

	serv := http.Server{
		Addr:    h.Address,
		Handler: proxy,
	}

	if launchdSocket == "" {
		return serv.ListenAndServe()
	}

	listeners, err := launch.SocketListeners(launchdSocket)
	if err != nil {
		return err
	}

	var t tomb.Tomb

	for _, l := range listeners {
		t.Go(func() error {
			return serv.Serve(l)
		})
	}

	return t.Wait()
}
