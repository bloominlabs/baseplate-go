package config

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/hashicorp/consul/sdk/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bloominlabs/baseplate-go/tlsutil"
)

func createTempCertificate(t *testing.T, filename string) (string, string) {
	certFile := testutil.TempFile(t, filename)
	pkFile := testutil.TempFile(t, filename)

	signer, _, err := tlsutil.GeneratePrivateKey()
	require.NoError(t, err)

	ca, _, err := tlsutil.GenerateCA(tlsutil.CAOpts{Signer: signer})
	require.NoError(t, err)

	cert, pk, err := tlsutil.GenerateCert(tlsutil.CertOpts{
		Signer:      signer,
		CA:          ca,
		Name:        "Test Cert Name",
		Days:        365,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	})
	require.NoError(t, err)

	_, err1 := certFile.WriteString(cert)
	_, err2 := pkFile.WriteString(pk)

	require.NoError(t, err1)
	require.NoError(t, err2)

	return certFile.Name(), pkFile.Name()
}

func TestNewCertificateWatcher(t *testing.T) {
	w, err := NewCertificateWatcher("", "", zerolog.Logger{}, 1*time.Nanosecond)
	require.NoError(t, err)
	require.NotNil(t, w)
}

func TestCertificateWatcherRenameEvent(t *testing.T) {
	t.Parallel()

	certFile1, pkFile1 := createTempCertificate(t, "set1")
	originalCertificate, err := tls.LoadX509KeyPair(certFile1, pkFile1)
	require.NoError(t, err)
	certFile2, pkFile2 := createTempCertificate(t, "set2")
	expectedCertificate, err := tls.LoadX509KeyPair(certFile1, pkFile1)
	require.NoError(t, err)

	w, err := NewCertificateWatcher(certFile1, pkFile1, log.With().Logger().Output(io.Discard), 1*time.Nanosecond)

	require.NoError(t, err)
	stop, err := w.Start(context.Background())
	require.NoError(t, err)
	defer stop()

	require.Equal(t, *w.cert, originalCertificate)
	err1 := os.Rename(certFile2, certFile1)
	err2 := os.Rename(pkFile2, pkFile1)
	require.NoError(t, err1)
	require.NoError(t, err2)
	require.Eventually(t, func() bool {
		return assert.Equal(t, *w.cert, expectedCertificate)
	},
		5*time.Second,
		50*time.Millisecond, "watcher did not rotate certificate within alotted time")
}

func TestCertificateWatcherStartNotCertificate(t *testing.T) {
	file := testutil.TempFile(t, "temp_config")
	filename := file.Name() + randomStr(16)
	_, err := NewCertificateWatcher(filename, filename, zerolog.Logger{}, 1*time.Nanosecond)
	require.Error(t, err, "no such file or directory")
}

func TestCertificateWatcherGetCertificate(t *testing.T) {
	t.Parallel()

	certFile1, pkFile1 := createTempCertificate(t, "set1")
	cert1, err := tls.LoadX509KeyPair(certFile1, pkFile1)
	require.NoError(t, err)
	certFile2, pkFile2 := createTempCertificate(t, "set2")
	cert2, err := tls.LoadX509KeyPair(certFile2, pkFile2)
	require.NoError(t, err)

	w, err := NewCertificateWatcher(certFile1, pkFile1, log.With().Logger(), 1*time.Nanosecond)

	require.NoError(t, err)
	stop, err := w.Start(context.Background())
	require.NoError(t, err)
	defer stop()

	// test the original certificate was picked up by the watcher
	getCert := w.GetCertificateFunc()
	require.Eventually(t, func() bool {
		actualCert, err := getCert(&tls.ClientHelloInfo{})
		require.NoError(t, err)
		return assert.Equal(t, cert1, *actualCert)
	},
		5*time.Second,
		500*time.Millisecond, "watcher did not rotate certificate within alotted time")

	// rotate the certificate on disk
	os.Rename(certFile2, certFile1)
	os.Rename(pkFile2, pkFile1)
	require.Eventually(t, func() bool {
		actualCert, err := getCert(&tls.ClientHelloInfo{})
		require.NoError(t, err)
		return assert.Equal(t, cert2, *actualCert)
	},
		5*time.Second,
		500*time.Millisecond, "watcher did not rotate certificate within alotted time")
}
