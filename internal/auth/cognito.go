// internal/auth/cognito.go
package auth

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v5"
)

type ctxKey string

const UserCtxKey ctxKey = "user"

// Estrutura básica dos claims do Cognito (access token)
type CognitoClaims struct {
	Scope    string `json:"scope,omitempty"`
	TokenUse string `json:"token_use,omitempty"` // "access" ou "id"
	ClientID string `json:"client_id,omitempty"`
	Username string `json:"username,omitempty"` // para access token
	Email    string `json:"email,omitempty"`    // no ID token
	jwt.RegisteredClaims
}

var jwks *keyfunc.JWKS

func getJWKSURL() string {
	region := os.Getenv("COGNITO_REGION")
	poolID := os.Getenv("COGNITO_USER_POOL_ID")
	return "https://cognito-idp." + region + ".amazonaws.com/" + poolID + "/.well-known/jwks.json"
}

func ensureJWKS() error {
	if jwks != nil {
		return nil
	}
	var err error
	jwks, err = keyfunc.Get(getJWKSURL(), keyfunc.Options{
		RefreshInterval:     time.Hour,
		RefreshErrorHandler: func(err error) { /* logar se quiser */ },
	})
	return err
}

func MiddlewareAutenticacaoCognito(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1) Pegar o token do header
		authz := r.Header.Get("Authorization")
		if authz == "" || !strings.HasPrefix(authz, "Bearer ") {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}
		raw := strings.TrimPrefix(authz, "Bearer ")

		// 2) JWKS
		if err := ensureJWKS(); err != nil {
			http.Error(w, "jwks unavailable", http.StatusInternalServerError)
			return
		}

		// 3) Validar assinatura e claims
		var claims CognitoClaims
		token, err := jwt.ParseWithClaims(raw, &claims, jwks.Keyfunc,
			jwt.WithAudience(os.Getenv("COGNITO_APP_CLIENT_ID")), // valida aud/client_id
			jwt.WithIssuer("https://cognito-idp."+os.Getenv("COGNITO_REGION")+
				".amazonaws.com/"+os.Getenv("COGNITO_USER_POOL_ID")),
		)
		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// 4) Checar tipo do token (recomendado usar ACCESS para APIs)
		if claims.TokenUse != "access" {
			http.Error(w, "wrong token type", http.StatusUnauthorized)
			return
		}

		// 5) Injetar identidade no contexto (username ou sub)
		ctx := context.WithValue(r.Context(), UserCtxKey, map[string]any{
			"sub":      claims.Subject,
			"username": claims.Username,
			"clientId": claims.ClientID,
			"scope":    claims.Scope,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper para pegar o usuário no handler
func GetUser(r *http.Request) (map[string]any, error) {
	u, ok := r.Context().Value(UserCtxKey).(map[string]any)
	if !ok || u == nil {
		return nil, errors.New("no user in context")
	}
	return u, nil
}
