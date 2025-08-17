package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims do seu token (inclui RBAC simples: IsAdmin)
type Claims struct {
	UserID  uint `json:"userId"`
	IsAdmin bool `json:"isAdmin"`
	jwt.RegisteredClaims
}

// Tempo de vida do access token
const AccessTTL = 15 * time.Minute

// Gera um JWT RS256 com KID, iss, aud, iat, nbf e jti

func GenerateAccessToken(userID uint, isAdmin bool) (string, error) {
	if err := mustInitKeys(); err != nil {
		return "", fmt.Errorf("keys init: %w", err)
	}
	priv := getPriv()
	if priv == nil {
		return "", fmt.Errorf("private key not loaded (check AUTH_RSA_PRIVATE_PATH and file permissions)")
	}

	now := time.Now()
	jti := fmt.Sprintf("%d-%d", userID, now.UnixNano())

	claims := &Claims{
		UserID:  userID,
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    getIssuer(),
			Audience:  []string{getAudience()},
			Subject:   fmt.Sprint(userID),
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-1 * time.Minute)),
			ID:        jti,
		},
	}

	tok := jwt.NewWithClaims(signMethod(), claims)
	tok.Header["kid"] = getKID()
	return tok.SignedString(priv)
}

// helper: verifica se a audience contém o valor esperado
func audienceContains(a jwt.ClaimStrings, want string) bool {
	for _, v := range a {
		if v == want {
			return true
		}
	}
	return false
}

// Valida assinatura, iss, aud e exp
func ParseAndValidate(tokenStr string) (*Claims, error) {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"RS256"}))
	tok, err := parser.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		k, _ := t.Header["kid"].(string)
		if k == "" {
			return nil, errors.New("kid ausente")
		}
		pub, ok := getPub(k)
		if !ok {
			return nil, errors.New("kid desconhecido")
		}
		return pub, nil
	})
	if err != nil {
		return nil, err
	}
	if !tok.Valid {
		return nil, errors.New("token inválido")
	}

	c, ok := tok.Claims.(*Claims)
	if !ok {
		return nil, errors.New("claims inválidas")
	}

	if c.Issuer != getIssuer() {
		return nil, errors.New("issuer inválido")
	}
	if !audienceContains(c.Audience, getAudience()) {
		return nil, errors.New("audience inválida")
	}
	if c.ExpiresAt == nil || time.Now().After(c.ExpiresAt.Time) {
		return nil, errors.New("token expirado")
	}

	return c, nil
}
