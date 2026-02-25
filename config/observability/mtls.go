package observability

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc/credentials"
)

func LoadKeyPair(caPath, certPath, keyPath string) (*credentials.TransportCredentials, error) {
	certificate, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	ca, err := os.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("can't read ca file from %s", caPath)
	}

	capool := x509.NewCertPool()
	if !capool.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("can't add CA cert to pool")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		RootCAs:      capool,
	}

	cred := credentials.NewTLS(tlsConfig)

	return &cred, nil
}
