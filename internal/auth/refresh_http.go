package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"time"

	"gorm.io/gorm"
)

const (
	RefreshTTL    = 30 * 24 * time.Hour
	RefreshCookie = "rt"
)

// --- Helpers ---

func genRaw() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashRaw(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// Em localhost (http://localhost) precisa ser Secure=false.
// Em produção (HTTPS), defina COOKIE_SECURE=true.
func cookieSecure() bool {
	return os.Getenv("COOKIE_SECURE") == "true"
}

func setRTCookie(w http.ResponseWriter, raw string, exp time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookie,
		Value:    raw,
		Path:     "/auth",              // cobre /auth/refresh e /auth/logout
		HttpOnly: true,
		Secure:   cookieSecure(),       // false no DEV (localhost), true em produção
		SameSite: http.SameSiteLaxMode, // ok p/ localhost:3000 -> localhost:8080
		Expires:  exp,
	})
}

func clearRTCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshCookie,
		Value:    "",
		Path:     "/auth",
		HttpOnly: true,
		Secure:   cookieSecure(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// --- Fluxo ---

// Use isso no LOGIN após validar usuário/senha
// isAdmin = true para Comercial (admin master); false para Consultor.
func IssueTokensOnLogin(db *gorm.DB, w http.ResponseWriter, userID uint, isAdmin bool) (string, error) {
	access, err := GenerateAccessToken(userID, isAdmin)
	if err != nil {
		return "", err
	}

	raw, err := genRaw()
	if err != nil {
		return "", err
	}

	rt := RefreshToken{
		UserID:    userID,
		FamilyID:  fmt.Sprintf("fam-%d", userID),
		Hash:      hashRaw(raw),
		IsAdmin:   isAdmin, // guarda o papel p/ RBAC no refresh
		ExpiresAt: time.Now().Add(RefreshTTL),
	}
	if err := db.Create(&rt).Error; err != nil {
		return "", err
	}
	setRTCookie(w, raw, rt.ExpiresAt)
	return access, nil
}

// POST /auth/refresh
func RefreshHTTPHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(RefreshCookie)
		if err != nil || c.Value == "" {
			http.Error(w, "no refresh", http.StatusUnauthorized)
			return
		}
		h := hashRaw(c.Value)

		var cur RefreshToken
		if err := db.Where("hash = ?", h).First(&cur).Error; err != nil {
			clearRTCookie(w)
			http.Error(w, "invalid refresh", http.StatusUnauthorized)
			return
		}
		if cur.RevokedAt != nil || time.Now().After(cur.ExpiresAt) {
			clearRTCookie(w)
			http.Error(w, "expired refresh", http.StatusUnauthorized)
			return
		}

		// revoke atual
		now := time.Now()
		_ = db.Model(&cur).Update("revoked_at", &now).Error

		// Gera novo access preservando RBAC do usuário salvo no refresh
		access, err := GenerateAccessToken(cur.UserID, cur.IsAdmin)
		if err != nil {
			clearRTCookie(w)
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}

		// novo refresh
		newRaw, err := genRaw()
		if err != nil {
			clearRTCookie(w)
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}
		newRT := RefreshToken{
			UserID:    cur.UserID,
			FamilyID:  cur.FamilyID,
			Hash:      hashRaw(newRaw),
			IsAdmin:   cur.IsAdmin, // mantém papel
			ExpiresAt: time.Now().Add(RefreshTTL),
		}
		if err := db.Create(&newRT).Error; err != nil {
			clearRTCookie(w)
			http.Error(w, "error", http.StatusInternalServerError)
			return
		}
		setRTCookie(w, newRaw, newRT.ExpiresAt)

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fmt.Sprintf(
			`{"access_token":"%s","token_type":"Bearer","expires_in":%d}`,
			access, int(AccessTTL.Seconds()),
		)))
	}
}

// POST /auth/logout
func LogoutHTTPHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if c, err := r.Cookie(RefreshCookie); err == nil && c.Value != "" {
			h := hashRaw(c.Value)
			now := time.Now()
			_ = db.Model(&RefreshToken{}).Where("hash = ?", h).Update("revoked_at", &now).Error
		}
		clearRTCookie(w)
		w.WriteHeader(http.StatusNoContent)
	}
}
