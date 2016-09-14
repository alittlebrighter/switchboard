package security

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/md5"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"io"
	"io/ioutil"
	"os"
)

var pubKeyCurve = elliptic.P256()

const privateKeyFile = "device.key"

func GenerateKeyPair() error {
	privateKey, err := ecdsa.GenerateKey(pubKeyCurve, rand.Reader)
	if err != nil {
		return err
	}

	privData, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return err
	}

	os.Remove(privateKeyFile)
	return ioutil.WriteFile(privateKeyFile, privData, 0400)
}

func fetchKeyPair() (*ecdsa.PrivateKey, error) {
	data, err := ioutil.ReadFile(privateKeyFile)
	if err != nil {
		return nil, err
	}

	return x509.ParseECPrivateKey(data)
}

func PublicKey() (ecdsa.PublicKey, error) {
	privKey, err := fetchKeyPair()
	if err != nil {
		return ecdsa.PublicKey{}, err
	}
	return privKey.PublicKey, nil
}

func Sign(msg string) (sig string, err error) {
	h := md5.New()
	_, err = io.WriteString(h, msg)
	if err != nil {
		return
	}
	hash := h.Sum(nil)

	privKey, err := fetchKeyPair()
	if err != nil {
		return
	}

	r, s, err := ecdsa.Sign(rand.Reader, privKey, hash)
	sig = base64.StdEncoding.EncodeToString(append(r.Bytes(), s.Bytes()...))
	return
}
