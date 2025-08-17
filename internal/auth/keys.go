package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/golang-jwt/jwt/v5"
)

var (
	keysOnce sync.Once
	keysErr  error

	privKey   *rsa.PrivateKey
	pubKeys   = map[string]*rsa.PublicKey{} // kid -> pub
	activeKID string
	issuer    string
	audience  string
)

func mustInitKeys() error {
	keysOnce.Do(func() {
		path := os.Getenv("AUTH_RSA_PRIVATE_PATH")
		activeKID = os.Getenv("AUTH_KID")
		issuer = os.Getenv("AUTH_ISSUER")
		audience = os.Getenv("AUTH_AUDIENCE")

		if path == "" || activeKID == "" || issuer == "" || audience == "" {
			keysErr = errors.New("missing envs: AUTH_RSA_PRIVATE_PATH/AUTH_KID/AUTH_ISSUER/AUTH_AUDIENCE")
			return
		}

		b, err := os.ReadFile(path)
		if err != nil {
			keysErr = fmt.Errorf("read private key: %w", err)
			return
		}
		block, _ := pem.Decode(b)
		if block == nil {
			keysErr = errors.New("pem decode private key failed")
			return
		}

		// PKCS#1 ou PKCS#8
		var pk any
		if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
			pk = k
		} else if k8, err2 := x509.ParsePKCS8PrivateKey(block.Bytes); err2 == nil {
			pk = k8
		} else {
			keysErr = fmt.Errorf("parse private key: %v / %v", err, err2)
			return
		}

		var ok bool
		privKey, ok = pk.(*rsa.PrivateKey)
		if !ok {
			keysErr = errors.New("private key is not RSA")
			return
		}
		pubKeys[activeKID] = &privKey.PublicKey
	})
	return keysErr
}

func getPriv() *rsa.PrivateKey                 { return privKey }
func getPub(kid string) (*rsa.PublicKey, bool) { p, ok := pubKeys[kid]; return p, ok }
func getKID() string                           { return activeKID }
func getIssuer() string                        { return issuer }
func getAudience() string                      { return audience }
func signMethod() jwt.SigningMethod            { return jwt.SigningMethodRS256 }
