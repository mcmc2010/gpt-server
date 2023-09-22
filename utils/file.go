package utils

import (
	"crypto/tls"
	"crypto/x509"
	"os"
)

// CERT Files
func LoadCertCAFromFile(filename string) *x509.CertPool {
	pem, err := os.ReadFile(filename)
	if err != nil {
		println("[Error] Load cert CA file (" + filename + ") error.")
		return nil
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		println("[Error] Load cert CA file (" + filename + ") error.")
		return nil
	}
	return pool
}

func LoadCertFromFiles(crt_filename string, key_filename string) *tls.Certificate {
	cert, err := tls.LoadX509KeyPair(crt_filename, key_filename)
	if err != nil {
		println("[Error] Load cert files (" + crt_filename + ", " + key_filename + ") error : " + err.Error())
		return nil
	}
	return &cert
}
