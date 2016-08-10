package dev

import (
	"crypto/tls"
	"net/http"
)

func (h *HTTPServer) ServeTLS() error {
	certCache := NewCertCache()

	tlsConfig := &tls.Config{
		GetCertificate: certCache.GetCertificate,
	}

	serv := http.Server{
		Addr:      h.TLSAddress,
		Handler:   h,
		TLSConfig: tlsConfig,
	}

	return serv.ListenAndServeTLS("", "")
}

func (h *HTTPServer) Serve() error {
	serv := http.Server{
		Addr:    h.Address,
		Handler: h,
	}

	return serv.ListenAndServe()
}
