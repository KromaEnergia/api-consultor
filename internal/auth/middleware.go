package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const UsuarioIDKey contextKey = "usuarioID"

// MiddlewareAutenticacao valida o Bearer token e injeta o userID no contexto.
func MiddlewareAutenticacao(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Permite preflight CORS
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Token ausente ou inválido", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := ValidarToken(tokenStr)
		if err != nil {
			http.Error(w, "Token inválido ou expirado", http.StatusUnauthorized)
			return
		}

		// Injeta o ID do consultor no contexto
		ctx := context.WithValue(r.Context(), UsuarioIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
