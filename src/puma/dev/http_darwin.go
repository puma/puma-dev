package dev

import (
	"crypto/tls"
	"net/http"
	"puma/dev/launch"
	"puma/httputil"
	"time"

	"gopkg.in/tomb.v2"
)

func (h *HTTPServer) ServeTLS(launchdSocket string) error {
	proxy := &httputil.ReverseProxy{
		Director:      h.director,
		Transport:     h.transport,
		FlushInterval: 1 * time.Second,
		Debug:         h.Debug,
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
		Debug:         h.Debug,
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
