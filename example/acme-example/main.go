// TODO https://github.com/cert-manager/cert-manager/issues/2131
package main

import (
	"context"
	"crypto/tls"
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/bloominlabs/baseplate-go/config"
	"github.com/rs/zerolog/log"
)

var tlsCertPath string
var tlsKeyPath string

func getenv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok == true {
		return val
	}

	return def
}

// TODO: make better desription
func init() {
	flag.StringVar(&tlsCertPath, "tls.cert.path", getenv("TLS_CERT_PATH", "cert.pem"), "TLS certificate to use for HTTPS server")
	flag.StringVar(&tlsKeyPath, "tls.key.path", getenv("TLS_KEY_PATH", "key.pem"), "TLS private key to use for HTTP server")
}

func main() {
	flag.Parse()
	log.Debug().Str("tlsCertPath", tlsCertPath).Str("tlsKeyPath", tlsKeyPath).Msg("starting acme-example")
	addr := ":" + os.Getenv("NOMAD_PORT_http")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world"))
	})

	w, err := config.NewCertificateWatcher(tlsCertPath, tlsKeyPath, log.Logger, time.Second*5)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create certificate watcher")
	}
	stop, err := w.Start(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start certificate watcher")
	}
	defer stop()

	server := &http.Server{
		Addr: addr,
		TLSConfig: &tls.Config{
			GetCertificate: w.GetCertificateFunc(),
		},
	}

	log.Debug().Str("tlsCertPath", tlsCertPath).Str("tlsKeyPath", tlsKeyPath).Str("addr", addr).Msg("starting https server")
	//Key and cert are coming from keypair reloader
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatal().Err(err).Msg("failed to listen and server TLS")
	}
}
