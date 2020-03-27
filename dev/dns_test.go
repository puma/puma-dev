package dev

import (
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

	tDNSResponder = NewDNSResponder("localhost:31337", domainList)

	go func() {
		if err := tDNSResponder.Serve(); err != nil {
			errChan <- err
		}
		close(errChan)
	}()

	shortTimeout := time.Duration(1 * time.Second)
	protocols := map[string](func() *dns.Server){
		"tcp": func() *dns.Server { return tDNSResponder.tcpServer },
		"udp": func() *dns.Server { return tDNSResponder.udpServer },
	}

	for protocol, serverLookup := range protocols {
		dialError := retry.Do(
			func() error {
				if _, err := net.DialTimeout(protocol, "localhost:31337", shortTimeout); err != nil {
					return err
				}

				defer func() {
					if server := serverLookup(); server != nil {
						server.Shutdown()
					} else {
						assert.Fail(t, "tDNSResponder", "%s was nil", protocol)
					}
				}()

				return nil
			},
		)

		assert.NoError(t, dialError)
	}

	assert.NoError(t, <-errChan)
}
