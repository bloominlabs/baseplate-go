package config

import (
	"context"
	"crypto/tls"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type CertificateWatcher struct {
	certMu   sync.RWMutex
	cert     *tls.Certificate
	certPath string
	keyPath  string
	watcher  Watcher
	logger   zerolog.Logger
}

func NewCertificateWatcher(certPath, keyPath string, logger zerolog.Logger, coalesceInterval time.Duration) (*CertificateWatcher, error) {
	w, err := NewRateLimitedFileWatcher([]string{certPath, keyPath}, logger, coalesceInterval)

	certWatcher := &CertificateWatcher{
		certPath: certPath,
		keyPath:  keyPath,
		logger:   logger,
		watcher:  w,
	}

	return certWatcher, err
}

func (w *CertificateWatcher) maybeReload() error {
	newCert, err := tls.LoadX509KeyPair(w.certPath, w.keyPath)
	if err != nil {
		return err
	}
	w.certMu.Lock()
	defer w.certMu.Unlock()
	w.cert = &newCert
	return nil
}

func (w *CertificateWatcher) Start(ctx context.Context) (func(), error) {
	w.watcher.Start(ctx)

	cert, err := tls.LoadX509KeyPair(w.certPath, w.keyPath)
	if err != nil {
		return nil, err
	}
	w.cert = &cert

	go func() {
		for event := range w.watcher.EventsCh() {
			w.logger.Debug().Int("num-events", len(event.Filenames)).Msg("certificate reload triggered")
			err := w.maybeReload()
			if err != nil {
				w.logger.Error().Err(err).Msg("error reloading certificates")
			}
		}
	}()

	return func() {
		w.watcher.Stop()
	}, nil
}

func (w *CertificateWatcher) Stop() error {
	return w.watcher.Stop()
}

func (w *CertificateWatcher) GetCertificateFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	if w.cert == nil {
		panic("did not find certificate in GetCertificateFunc() call. did you run watcher.Start()? Start() is necessary to ensure goroutines are started and cleaned up properly")
	}

	return func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		w.certMu.RLock()
		defer w.certMu.RUnlock()
		return w.cert, nil
	}
}
