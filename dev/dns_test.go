package dev

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/avast/retry-go"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
)

var tDNSResponder *DNSResponder

func TestServeDNS(t *testing.T) {
	errChan := make(chan error, 1)
	domainList := []string{"test"}
	dnsAddress := "localhost:10053"
	shortTimeout := time.Duration(1 * time.Second)

	tDNSResponder = NewDNSResponder(dnsAddress, domainList)

	go func() {
		defer close(errChan)
		if err := tDNSResponder.Serve(); err != nil {
			errChan <- err
		}
	}()

	protocols := map[string]*dns.Server{
		"tcp": tDNSResponder.tcpServer,
		"udp": tDNSResponder.udpServer,
	}

	for protocol, server := range protocols {
		dialError := retry.Do(
			func() error {
				if _, err := net.DialTimeout(protocol, dnsAddress, shortTimeout); err != nil {
					return err
				}

				assert.NoError(t, server.Shutdown())

				return nil
			},
			retry.Attempts(3),
		)

		assert.NoError(t, dialError)
	}

	responderError := <-errChan
	if responderError != nil {
		fmt.Printf("%+v", responderError)
	}
	assert.NoError(t, responderError)
}
