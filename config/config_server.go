package config

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type ServerConfig struct {
	Address  string
	CertPath string `toml:"cert_path"`
	KeyPath  string `toml:"key_path"`

	watcher *CertificateWatcher
}

func (c *ServerConfig) RegisterFlags(f *flag.FlagSet, prefix string) {
	flag.StringVar(&c.Address, fmt.Sprintf("%s.addr", prefix), getenv("NOMAD_ADDR_server", ":8080"), "hostname:port to connect to server")

	flag.StringVar(&c.CertPath, fmt.Sprintf("%s.tls.cert.path", prefix), "", "Path to the TLS certificate file")
	flag.StringVar(&c.KeyPath, fmt.Sprintf("%s.tls.key.path", prefix), "", "Path to the TLS key file")
}

func (c *ServerConfig) NewServer(mux *http.ServeMux, logger zerolog.Logger) (*http.Server, error) {
	mux.HandleFunc("/.well-known/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:              c.Address,
		Handler:           mux,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}
	if c.CertPath != "" || c.KeyPath != "" {
		w, err := NewCertificateWatcher(c.CertPath, c.KeyPath, logger, time.Second*5)
		if err != nil {

			logger.Fatal().Err(err).Msg("failed to create certificate watcher")
		}
		_, err = w.Start(context.Background())
		if err != nil {
			return nil, err
		}
		c.watcher = w
		server.TLSConfig = &tls.Config{
			GetCertificate: w.GetCertificateFunc(),
		}
	} else {
		logger.Warn().Msg("tls certificate path and key path are not specified. using http instead of https")
	}

	return server, nil
}

func (c *ServerConfig) Shutdown(ctx context.Context, logger zerolog.Logger) error {
	if c.watcher != nil {
		return c.watcher.Stop()
	}

	return nil
}
