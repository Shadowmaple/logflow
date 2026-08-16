package utils

import (
	"crypto/x509"
	"fmt"
	"os"
)

// loadCACert 加载CA根证书文件
func LoadCACert(caFilePath string) (*x509.CertPool, error) {
	caPEM, err := os.ReadFile(caFilePath)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("failed to append ca cert")
	}
	return pool, nil
}
