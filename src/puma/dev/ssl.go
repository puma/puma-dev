package dev

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

var CACert *tls.Certificate

func SetupOurCert() error {
	dir := mustExpand("~/Library/Application Support/io.puma.dev")

	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return err
	}

	key := filepath.Join(dir, "key.pem")
	cert := filepath.Join(dir, "cert.pem")

	tlsCert, err := tls.LoadX509KeyPair(cert, key)
	if err == nil {
		CACert = &tlsCert
		return nil
	}

	err = exec.Command("sh", "-c",
		fmt.Sprintf(
			`openssl req -newkey rsa:2048 -batch -x509 -sha256 -nodes -subj "/C=US/O=Developer Certificate/CN=Puma-dev CA" -keyout '%s' -out '%s' -days 9999`,
			key, cert)).Run()
	if err != nil {
		return err
	}

	fmt.Printf("* Adding certification to login keychain as trusted\n")
	fmt.Printf("! There is probably a dialog open that you must type your password into\n")

	login := mustExpand("~/Library/Keychains/login.keychain")

	err = exec.Command("sh", "-c",
		fmt.Sprintf(`security add-trusted-cert -d -r trustRoot -k '%s' '%s'`,
			login, cert)).Run()

	if err != nil {
		return err
	}

	fmt.Printf("* Certificates setup, ready for https operations!\n")
	return nil
}

type certCache struct {
	lock  sync.Mutex
	cache map[string]*tls.Certificate
}

func NewCertCache() *certCache {
	return &certCache{
		cache: make(map[string]*tls.Certificate),
	}
}

func (c *certCache) GetCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	name := clientHello.ServerName

	if cert, ok := c.cache[name]; ok {
		return cert, nil
	}

	cert, err := makeCert(CACert, name)
	if err != nil {
		return nil, err
	}

	c.cache[name] = cert

	return cert, nil
}

func makeCert(
	parent *tls.Certificate,
	name string,
) (*tls.Certificate, error) {

	// start by generating private key
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	// create certificate structure with proper values
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %v", err)
	}

	cert := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Puma-dev Signed"},
			CommonName:   name,
		},
		NotBefore:   notBefore,
		NotAfter:    notAfter,
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	cert.DNSNames = append(cert.DNSNames, name)

	x509parent, err := x509.ParseCertificate(parent.Certificate[0])
	if err != nil {
		return nil, err
	}

	derBytes, err := x509.CreateCertificate(
		rand.Reader, cert, x509parent, privKey.Public(), parent.PrivateKey)

	if err != nil {
		return nil, fmt.Errorf("could not create certificate: %v", err)
	}

	tlsCert := &tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  privKey,
		Leaf:        cert,
	}

	return tlsCert, nil
}
