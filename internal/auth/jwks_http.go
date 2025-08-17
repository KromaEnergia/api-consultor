package auth

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
)

type jwk struct {
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// GET /.well-known/jwks.json
func JWKSHandler(w http.ResponseWriter, r *http.Request) {
	if err := mustInitKeys(); err != nil {
		http.Error(w, "jwks unavailable", http.StatusInternalServerError)
		return
	}
	pub, ok := getPub(getKID())
	if !ok || pub == nil {
		http.Error(w, "no public key", http.StatusInternalServerError)
		return
	}

	n := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes())

	resp := struct {
		Keys []jwk `json:"keys"`
	}{
		Keys: []jwk{{
			Kty: "RSA",
			Alg: "RS256",
			Use: "sig",
			Kid: getKID(),
			N:   n,
			E:   e,
		}},
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
