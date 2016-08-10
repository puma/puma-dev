package dev

import (
	"crypto/tls"
	"net/http"

	"github.com/puma/puma-dev/dev/launch"

	"gopkg.in/tomb.v2"
)

func (h *HTTPServer) ServeTLS(launchdSocket string) error {
	certCache := NewCertCache()

	tlsConfig := &tls.Config{
		GetCertificate: certCache.GetCertificate,
	}

	serv := http.Server{
		Addr:      h.TLSAddress,
		Handler:   h,
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
	serv := http.Server{
		Addr:    h.Address,
		Handler: h,
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
