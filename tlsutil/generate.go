package tlsutil

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"strings"
	"time"

	"net/url"
)

const (
	DefaultPrivateKeyType      = "ec"
	DefaultPrivateKeyBits      = 256
	DefaultIntermediateCertTTL = 24 * 365 * time.Hour
)

func pemEncode(value []byte, blockType string) (string, error) {
	var buf bytes.Buffer

	if err := pem.Encode(&buf, &pem.Block{Type: blockType, Bytes: value}); err != nil {
		return "", fmt.Errorf("error encoding value %v: %s", blockType, err)
	}
	return buf.String(), nil
}

// ParseSigner parses a crypto.Signer from a PEM-encoded key. The private key
// is expected to be the first block in the PEM value.
func ParseSigner(pemValue string) (crypto.Signer, error) {
	// The _ result below is not an error but the remaining PEM bytes.
	block, _ := pem.Decode([]byte(pemValue))
	if block == nil {
		return nil, fmt.Errorf("no PEM-encoded data found")
	}

	switch block.Type {
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(block.Bytes)

	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)

	case "PRIVATE KEY":
		signer, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		pk, ok := signer.(crypto.Signer)
		if !ok {
			return nil, fmt.Errorf("private key is not a valid format")
		}

		return pk, nil

	default:
		return nil, fmt.Errorf("unknown PEM block type for signing key: %s", block.Type)
	}
}

// GenerateSerialNumber returns random bigint generated with crypto/rand
func GenerateSerialNumber() (*big.Int, error) {
	l := new(big.Int).Lsh(big.NewInt(1), 128)
	s, err := rand.Int(rand.Reader, l)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func generateECDSAKey(keyBits int) (crypto.Signer, string, error) {
	var pk *ecdsa.PrivateKey
	var curve elliptic.Curve

	switch keyBits {
	case 224:
		curve = elliptic.P224()
	case 256:
		curve = elliptic.P256()
	case 384:
		curve = elliptic.P384()
	case 521:
		curve = elliptic.P521()
	default:
		return nil, "", fmt.Errorf("error generating ECDSA private key: unknown curve length %d", keyBits)
	}

	pk, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, "", fmt.Errorf("error generating ECDSA private key: %s", err)
	}

	bs, err := x509.MarshalECPrivateKey(pk)
	if err != nil {
		return nil, "", fmt.Errorf("error marshaling ECDSA private key: %s", err)
	}

	pemBlock, err := pemEncode(bs, "EC PRIVATE KEY")
	if err != nil {
		return nil, "", err
	}

	return pk, pemBlock, nil
}

func generateRSAKey(keyBits int) (crypto.Signer, string, error) {
	var pk *rsa.PrivateKey

	pk, err := rsa.GenerateKey(rand.Reader, keyBits)
	if err != nil {
		return nil, "", fmt.Errorf("error generating RSA private key: %s", err)
	}

	bs := x509.MarshalPKCS1PrivateKey(pk)
	pemBlock, err := pemEncode(bs, "RSA PRIVATE KEY")
	if err != nil {
		return nil, "", err
	}

	return pk, pemBlock, nil
}

// GeneratePrivateKey generates a new Private key
func GeneratePrivateKeyWithConfig(keyType string, keyBits int) (crypto.Signer, string, error) {
	switch strings.ToLower(keyType) {
	case "rsa":
		return generateRSAKey(keyBits)
	case "ec":
		return generateECDSAKey(keyBits)
	default:
		return nil, "", fmt.Errorf("unknown private key type requested: %s", keyType)
	}
}

func GeneratePrivateKey() (crypto.Signer, string, error) {
	// TODO: find any calls to this func, replace with calls to GeneratePrivateKeyWithConfig()
	// using prefs `private_key_type` and `private_key_bits`
	return GeneratePrivateKeyWithConfig(DefaultPrivateKeyType, DefaultPrivateKeyBits)
}

type CAOpts struct {
	Signer              crypto.Signer
	Serial              *big.Int
	ClusterID           string
	Days                int
	PermittedDNSDomains []string
	Domain              string
	Name                string
}

type CertOpts struct {
	Signer      crypto.Signer
	CA          string
	Serial      *big.Int
	Name        string
	Days        int
	DNSNames    []string
	IPAddresses []net.IP
	ExtKeyUsage []x509.ExtKeyUsage
	IsCA        bool
}

// GenerateCA generates a new CA for agent TLS (not to be confused with Connect TLS)
func GenerateCA(opts CAOpts) (string, string, error) {
	signer := opts.Signer
	var pk string
	if signer == nil {
		var err error
		signer, pk, err = GeneratePrivateKey()
		if err != nil {
			return "", "", err
		}
	}

	id, err := keyID(signer.Public())
	if err != nil {
		return "", "", err
	}

	sn := opts.Serial
	if sn == nil {
		var err error
		sn, err = GenerateSerialNumber()
		if err != nil {
			return "", "", err
		}
	}
	name := opts.Name
	if name == "" {
		name = fmt.Sprintf("Consul Agent CA %d", sn)
	}

	days := opts.Days
	if opts.Days == 0 {
		days = 365
	}

	var uris []*url.URL

	// Create the CA cert
	template := x509.Certificate{
		SerialNumber: sn,
		URIs:         uris,
		Subject: pkix.Name{
			Organization: []string{"Bloomin' Labs, LLC"},
			CommonName:   name,
		},
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		IsCA:                  true,
		NotAfter:              time.Now().AddDate(0, 0, days),
		NotBefore:             time.Now(),
		AuthorityKeyId:        id,
		SubjectKeyId:          id,
	}

	if len(opts.PermittedDNSDomains) > 0 {
		template.PermittedDNSDomainsCritical = true
		template.PermittedDNSDomains = opts.PermittedDNSDomains
	}
	bs, err := x509.CreateCertificate(
		rand.Reader, &template, &template, signer.Public(), signer)
	if err != nil {
		return "", "", fmt.Errorf("error generating CA certificate: %s", err)
	}

	var buf bytes.Buffer
	err = pem.Encode(&buf, &pem.Block{Type: "CERTIFICATE", Bytes: bs})
	if err != nil {
		return "", "", fmt.Errorf("error encoding private key: %s", err)
	}

	return buf.String(), pk, nil
}

// GenerateCert generates a new certificate for TLS
func GenerateCert(opts CertOpts) (string, string, error) {
	parent, err := parseCert(opts.CA)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse CA: %w", err)
	}

	signee, pk, err := GeneratePrivateKey()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}

	id, err := keyID(signee.Public())
	if err != nil {
		return "", "", fmt.Errorf("failed to get keyID from public key: %w", err)
	}

	sn := opts.Serial
	if sn == nil {
		var err error
		sn, err = GenerateSerialNumber()
		if err != nil {
			return "", "", err
		}
	}

	template := x509.Certificate{
		SerialNumber:          sn,
		Subject:               pkix.Name{CommonName: opts.Name},
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           opts.ExtKeyUsage,
		IsCA:                  false,
		NotAfter:              time.Now().AddDate(0, 0, opts.Days),
		NotBefore:             time.Now(),
		SubjectKeyId:          id,
		DNSNames:              opts.DNSNames,
		IPAddresses:           opts.IPAddresses,
	}
	if opts.IsCA {
		template.IsCA = true
		template.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature
	}

	bs, err := x509.CreateCertificate(rand.Reader, &template, parent, signee.Public(), opts.Signer)
	if err != nil {
		return "", "", fmt.Errorf("failed to create certificate: %w", err)
	}

	var buf bytes.Buffer
	err = pem.Encode(&buf, &pem.Block{Type: "CERTIFICATE", Bytes: bs})
	if err != nil {
		return "", "", fmt.Errorf("error encoding private key: %s", err)
	}

	return buf.String(), pk, nil
}

// KeyId returns a x509 KeyId from the given signing key.
func keyID(raw interface{}) ([]byte, error) {
	switch raw.(type) {
	case *ecdsa.PublicKey:
	case *rsa.PublicKey:
	default:
		return nil, fmt.Errorf("invalid key type: %T", raw)
	}

	// This is not standard; RFC allows any unique identifier as long as they
	// match in subject/authority chains but suggests specific hashing of DER
	// bytes of public key including DER tags.
	bs, err := x509.MarshalPKIXPublicKey(raw)
	if err != nil {
		return nil, err
	}

	// String formatted
	kID := sha256.Sum256(bs)
	return kID[:], nil
}

func parseCert(pemValue string) (*x509.Certificate, error) {
	// The _ result below is not an error but the remaining PEM bytes.
	block, _ := pem.Decode([]byte(pemValue))
	if block == nil {
		return nil, fmt.Errorf("no PEM-encoded data found")
	}

	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("first PEM-block should be CERTIFICATE type")
	}

	return x509.ParseCertificate(block.Bytes)
}

func Verify(caString, certString, dns string) error {
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(caString))
	if !ok {
		return fmt.Errorf("failed to parse root certificate")
	}

	cert, err := parseCert(certString)
	if err != nil {
		return fmt.Errorf("failed to parse certificate")
	}

	opts := x509.VerifyOptions{
		DNSName: fmt.Sprint(dns),
		Roots:   roots,
	}

	_, err = cert.Verify(opts)
	return err
}
