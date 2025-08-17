package auth

import (
	"context"
	"net/http"
	"strings"
)

type ctxKey string
const (
	CtxUserID  ctxKey = "usuarioID"
	CtxIsAdmin ctxKey = "isAdmin"
)

func MiddlewareAutenticacao(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		h := r.Header.Get("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			http.Error(w, "Token ausente", http.StatusUnauthorized); return
		}
		raw := strings.TrimPrefix(h, "Bearer ")
		claims, err := ParseAndValidate(raw)
		if err != nil {
			http.Error(w, "Token inválido", http.StatusUnauthorized); return
		}
		ctx := context.WithValue(r.Context(), CtxUserID, claims.UserID)
		ctx = context.WithValue(ctx, CtxIsAdmin, claims.IsAdmin)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := r.Context().Value(CtxIsAdmin)
		if ok, _ := v.(bool); !ok {
			http.Error(w, "Forbidden (admin only)", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
