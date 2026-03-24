package bootstrap

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	fakewebhook "payrune/internal/adapters/inbound/http/fakewebhook"
)

const defaultFakeWebhookReceiverAddr = ":8443"

type FakeWebhookReceiverConfig struct {
	Addr   string
	Secret string
	Logger *log.Logger
}

func LoadFakeWebhookReceiverConfigFromEnv() FakeWebhookReceiverConfig {
	return loadFakeWebhookReceiverConfigFromLookup(os.Getenv)
}

func RunFakeWebhookReceiver(config FakeWebhookReceiverConfig) error {
	if strings.TrimSpace(config.Addr) == "" {
		config.Addr = defaultFakeWebhookReceiverAddr
	}

	certificate, err := generateSelfSignedCertificate(time.Now())
	if err != nil {
		return err
	}

	logger := config.Logger
	if logger == nil {
		logger = log.Default()
	}

	server := &http.Server{
		Addr:    config.Addr,
		Handler: fakewebhook.NewHandler(logger, config.Secret),
		TLSConfig: &tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{certificate},
		},
	}

	listener, err := tls.Listen("tcp", config.Addr, server.TLSConfig)
	if err != nil {
		return err
	}

	logger.Printf(
		"fake webhook receiver listening on %s verification_method=hmac-sha256 verification_enabled=%t",
		config.Addr,
		config.Secret != "",
	)
	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func loadFakeWebhookReceiverConfigFromLookup(lookup func(string) string) FakeWebhookReceiverConfig {
	addr := strings.TrimSpace(lookup("FAKE_WEBHOOK_RECEIVER_ADDR"))
	if addr == "" {
		addr = defaultFakeWebhookReceiverAddr
	}

	return FakeWebhookReceiverConfig{
		Addr:   addr,
		Secret: loadFakeWebhookReceiverSecret(lookup),
	}
}

func loadFakeWebhookReceiverSecret(lookup func(string) string) string {
	if secret := strings.TrimSpace(lookup("FAKE_WEBHOOK_RECEIVER_SECRET")); secret != "" {
		return secret
	}
	return strings.TrimSpace(lookup("PAYMENT_RECEIPT_WEBHOOK_SECRET"))
}

func generateSelfSignedCertificate(now time.Time) (tls.Certificate, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "receipt-webhook-mock",
		},
		NotBefore:             now.Add(-1 * time.Hour),
		NotAfter:              now.Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames: []string{
			"localhost",
			"receipt-webhook-mock",
			"receipt-webhook-dispatcher",
		},
		IPAddresses: []net.IP{
			net.ParseIP("127.0.0.1"),
		},
	}

	certificateDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}

	certificatePEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificateDER})
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	return tls.X509KeyPair(certificatePEM, privateKeyPEM)
}
