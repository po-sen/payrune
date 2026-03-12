package main

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
	"time"
)

func main() {
	certificate, err := generateSelfSignedCertificate()
	if err != nil {
		log.Fatalf("generate self-signed certificate: %v", err)
	}

	address := ":8443"
	if raw := os.Getenv("FAKE_WEBHOOK_RECEIVER_ADDR"); raw != "" {
		address = raw
	}
	secret := loadFakeWebhookReceiverSecret()
	logger := log.Default()

	server := &http.Server{
		Addr:    address,
		Handler: newFakeWebhookHandler(logger, secret),
		TLSConfig: &tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: []tls.Certificate{certificate},
		},
	}

	listener, err := tls.Listen("tcp", address, server.TLSConfig)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	logger.Printf(
		"fake webhook receiver listening on %s verification_method=hmac-sha256 verification_enabled=%t",
		address,
		secret != "",
	)
	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		log.Fatalf("serve: %v", err)
	}
}

func loadFakeWebhookReceiverSecret() string {
	if secret := os.Getenv("FAKE_WEBHOOK_RECEIVER_SECRET"); secret != "" {
		return secret
	}
	return os.Getenv("PAYMENT_RECEIPT_WEBHOOK_SECRET")
}

func generateSelfSignedCertificate() (tls.Certificate, error) {
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
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
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
