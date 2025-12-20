package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/quic-go/quic-go/http3"
)

// setupTLSConfig настраивает TLS конфигурацию для сервера
func (s *Server) setupTLSConfig() (*tls.Config, error) {
	// Загружаем сертификат сервера
	cert, err := s.loadServerCertificate()
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate: %w", err)
	}

	// Загружаем CA сертификат для проверки клиентов
	caCertPool, err := s.loadCACertificate()
	if err != nil {
		return nil, fmt.Errorf("failed to load CA certificate: %w", err)
	}

	// Создаем TLS конфигурацию
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caCertPool,
		NextProtos:   []string{http3.NextProtoH3},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		},
	}

	return tlsConfig, nil
}

// loadServerCertificate загружает сертификат сервера
func (s *Server) loadServerCertificate() (tls.Certificate, error) {
	// Приоритет: PEM из конфигурации, затем файлы
	if s.Config.CertPEM != "" && s.Config.KeyPEM != "" {
		cert, err := tls.X509KeyPair([]byte(s.Config.CertPEM), []byte(s.Config.KeyPEM))
		if err != nil {
			return tls.Certificate{}, fmt.Errorf("failed to load certificate from PEM config: %w", err)
		}
		return cert, nil
	}

	if s.Config.CertFile == "" || s.Config.KeyFile == "" {
		return tls.Certificate{}, fmt.Errorf("certificate and key files must be specified")
	}

	cert, err := tls.LoadX509KeyPair(s.Config.CertFile, s.Config.KeyFile)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to load certificate from files: %w", err)
	}

	return cert, nil
}

// loadCACertificate загружает CA сертификат для проверки клиентов
func (s *Server) loadCACertificate() (*x509.CertPool, error) {
	caCertPool := x509.NewCertPool()

	var caCert []byte
	var err error

	// Приоритет: PEM из конфигурации, затем файл
	if s.Config.CACertPEM != "" {
		caCert = []byte(s.Config.CACertPEM)
	} else {
		if s.Config.CACertFile == "" {
			return nil, fmt.Errorf("CA certificate file or PEM must be specified for mutual TLS")
		}

		caCert, err = os.ReadFile(s.Config.CACertFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate file %s: %w", s.Config.CACertFile, err)
		}
	}

	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	return caCertPool, nil
}

// validateClientCertificate проверяет клиентский сертификат
func (s *Server) validateClientCertificate(cert *x509.Certificate) error {
	if cert == nil {
		return fmt.Errorf("client certificate is required")
	}

	// Проверяем Common Name
	if cert.Subject.CommonName == "" {
		return fmt.Errorf("client certificate must have Common Name")
	}

	// Проверяем срок действия
	now := cert.NotAfter
	if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
		return fmt.Errorf("client certificate is expired or not yet valid")
	}

	// Дополнительные проверки можно добавить здесь
	// Например, проверка по базе данных разрешенных клиентов

	return nil
}