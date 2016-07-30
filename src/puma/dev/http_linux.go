package dev

import (
	"crypto/tls"
	"net/http"
	"puma/httputil"
	"time"
)

func (h *HTTPServer) ServeTLS() error {
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

	return serv.ListenAndServeTLS("", "")
}

func (h *HTTPServer) Serve() error {
	proxy := &httputil.ReverseProxy{
		Director:      h.director,
		Transport:     h.transport,
		FlushInterval: 1 * time.Second,
	}

	serv := http.Server{
		Addr:    h.Address,
		Handler: proxy,
	}

	return serv.ListenAndServe()
}
