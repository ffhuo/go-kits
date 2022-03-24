package x509cert

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

func Generate(expTime time.Duration, ip []string) (certbuf *bytes.Buffer, keybuf *bytes.Buffer, err error) {
	max := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNum, _ := rand.Int(rand.Reader, max)

	template := x509.Certificate{
		SerialNumber: serialNum,
		Subject:      pkix.Name{CommonName: fmt.Sprintf("JYSSO@%d", time.Now().Unix())},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(expTime),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}
	if len(ip) > 0 {
		for _, v := range ip {
			template.IPAddresses = append(template.IPAddresses, net.ParseIP(v))
		}
	}
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return
	}

	cert, err := x509.CreateCertificate(rand.Reader, &template, &template, &pk.PublicKey, pk)
	if err != nil {
		return
	}

	certbuf = new(bytes.Buffer)
	pem.Encode(certbuf, &pem.Block{Type: "CERTIFICATE", Bytes: cert})

	keybuf = new(bytes.Buffer)
	key := x509.MarshalPKCS1PrivateKey(pk)
	pem.Encode(keybuf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: key})
	return
}
