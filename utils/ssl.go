package utils

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
)

// ConfigureSecureHTTPClientWithFile creates an HTTP client with certificates from a file
func ConfigureSecureHTTPClientWithFile(certPath string) (*http.Client, error) {
	if certPath == "" {
		return nil, nil
	}

	// Create a new certificate pool
	rootCAs, err := x509.SystemCertPool()
	if err != nil || rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	// Read the certificate file
	certsData, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("could not read certificate file: %w", err)
	}

	// Add the certificates to the pool
	if ok := rootCAs.AppendCertsFromPEM(certsData); !ok {
		return nil, fmt.Errorf("could not add certificates to the pool")
	}

	// Create an HTTP client with the custom configuration
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: rootCAs,
			},
		},
	}

	return client, nil
}