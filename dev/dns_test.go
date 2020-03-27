package dev

import (
	"log"
	"net"
	"testing"
	"time"

	"github.com/avast/retry-go"
	"github.com/stretchr/testify/assert"
)

var tDNSResponder DNSResponder

func TestServe(t *testing.T) {
	tDNSResponder.Address = "localhost:31337"
	errChan := make(chan error)

	go func() {
		if err := tDNSResponder.Serve([]string{"test"}); err != nil {
			errChan <- err
		}
		close(errChan)
	}()

	tcpErr := retry.Do(
		func() error {
			if _, err := net.DialTimeout("tcp", "localhost:31337", time.Duration(10*time.Second)); err != nil {
				return err
			}
			tDNSResponder.tcpServer.Shutdown()
			return nil
		},
		retry.OnRetry(func(n uint, err error) {
			log.Printf("#%d: %s\n", n, err)
		}),
	)

	udpErr := retry.Do(
		func() error {
			if _, err := net.DialTimeout("udp", "localhost:31337", time.Duration(10*time.Second)); err != nil {
				return err
			}
			tDNSResponder.udpServer.Shutdown()
			return nil
		},
		retry.OnRetry(func(n uint, err error) {
			log.Printf("#%d: %s\n", n, err)
		}),
	)

	assert.NoError(t, tcpErr)
	assert.NoError(t, udpErr)
}
