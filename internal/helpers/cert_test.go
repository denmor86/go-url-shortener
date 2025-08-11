package helpers

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSelfSignedCert(t *testing.T) {
	t.Run("Successful", func(t *testing.T) {
		certPEM, privKey, err := GenerateSelfSignedCert()
		require.NoError(t, err)
		require.NotNil(t, certPEM)
		require.NotNil(t, privKey)

		// Проверяем PEM кодировку сертификата
		block, _ := pem.Decode(certPEM)
		require.NotNil(t, block, "should decode PEM block successfully")
		assert.Equal(t, "CERTIFICATE", block.Type, "PEM block should be of type CERTIFICATE")

		// Парсим сертификат
		cert, err := x509.ParseCertificate(block.Bytes)
		require.NoError(t, err)

		// Проверяем основные поля сертификата
		assert.Equal(t, big.NewInt(1658), cert.SerialNumber)
		assert.Equal(t, "Yandex.Praktikum", cert.Subject.Organization[0])
		assert.Equal(t, "RU", cert.Subject.Country[0])
		assert.True(t, time.Now().After(cert.NotBefore), "certificate should be valid now")
		assert.True(t, time.Now().AddDate(10, 0, 0).Sub(cert.NotAfter) < time.Hour*24, "certificate should expire in ~10 years")

		// Проверяем IP адреса
		expectedIPs := []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}
		require.Len(t, cert.IPAddresses, len(expectedIPs), "количество IP адресов не совпадает")

		for _, expectedIP := range expectedIPs {
			found := false
			for _, actualIP := range cert.IPAddresses {
				if expectedIP.Equal(actualIP) {
					found = true
					break
				}
			}
			assert.True(t, found, "IP %v не найден в сертификате", expectedIP)
		}

		// Проверяем KeyUsage
		assert.Equal(t, x509.KeyUsageDigitalSignature, cert.KeyUsage)
		assert.ElementsMatch(t, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}, cert.ExtKeyUsage)

		// Проверяем соответствие публичного ключа
		pubKey, ok := cert.PublicKey.(*rsa.PublicKey)
		require.True(t, ok, "public key should be RSA")
		assert.Equal(t, privKey.PublicKey.N, pubKey.N, "public key modulus should match")
		assert.Equal(t, privKey.PublicKey.E, pubKey.E, "public key exponent should match")
	})

	t.Run("Check private key", func(t *testing.T) {
		_, privKey, err := GenerateSelfSignedCert()
		require.NoError(t, err)

		// Проверяем размер ключа (4096 бит)
		assert.Equal(t, 4096, privKey.Size()*8, "key size should be 4096 bits")

		// Проверяем, что ключ можно использовать для подписи
		err = privKey.Validate()
		assert.NoError(t, err, "private key should be valid")
	})
}
