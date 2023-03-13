package server

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/rs/zerolog"

	"github.com/bloominlabs/baseplate-go/config/env"
	"github.com/bloominlabs/baseplate-go/config/filesystem"
)

type ServerConfig struct {
	Address  string
	CertPath string `toml:"cert_path"`
	KeyPath  string `toml:"key_path"`
}

func (c *ServerConfig) RegisterFlags(f *flag.FlagSet, prefix string) {
	f.StringVar(&c.Address, fmt.Sprintf("%s.addr", prefix), env.GetEnvStrDefault(fmt.Sprintf("NOMAD_ADDR_%s", prefix), ":8080"), "hostname:port to connect to server")

	f.StringVar(&c.CertPath, fmt.Sprintf("%s.tls.cert.path", prefix), "", "Path to the TLS certificate file")
	f.StringVar(&c.KeyPath, fmt.Sprintf("%s.tls.key.path", prefix), "", "Path to the TLS key file")
}

func (c *ServerConfig) UseCommonRoutes(mux *http.ServeMux, public bool) {
	mux.HandleFunc("/.well-known/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	if !public {
		// handling pprof
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
		mux.HandleFunc("/debug/pprof/", pprof.Index)
	}
}

func (c *ServerConfig) NewServer(handler http.Handler, logger zerolog.Logger) (*Server, error) {
	server := http.Server{
		Addr:              c.Address,
		Handler:           handler,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}
	var watcher *filesystem.CertificateWatcher
	if c.CertPath != "" || c.KeyPath != "" {
		w, err := filesystem.NewCertificateWatcher(c.CertPath, c.KeyPath, logger, time.Second*5)
		if err != nil {

			logger.Fatal().Err(err).Msg("failed to create certificate watcher")
		}
		_, err = w.Start(context.Background())
		if err != nil {
			return nil, err
		}
		watcher = w
		server.TLSConfig = &tls.Config{
			GetCertificate: w.GetCertificateFunc(),
		}
	} else {
		logger.Warn().Msg("tls certificate path and key path are not specified. using http instead of https")
	}

	return &Server{
		server,
		logger,
		watcher,
	}, nil
}

type Server struct {
	http.Server

	logger  zerolog.Logger
	watcher *filesystem.CertificateWatcher
}

func (c *Server) Listen() error {
	if c.TLSConfig != nil {
		if err := c.ListenAndServeTLS("", ""); err != nil {
			return fmt.Errorf("error while serving server: %w", err)
		}
	} else {
		c.logger.Warn().Msg("running http server without https. this is not in production")
		if err := c.ListenAndServe(); err != nil {
			return fmt.Errorf("error while serving server: %w", err)
		}
	}

	return nil
}

func (c *Server) Cleanup() error {
	if c.watcher != nil {
		return c.watcher.Stop()
	}

	return nil
}
