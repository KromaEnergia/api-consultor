package auth

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

var jwtSecret []byte

func init() {
	// Carrega variáveis do .env (ou tenta usar do ambiente)
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: .env não encontrado, usando variáveis de ambiente do sistema")
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("Erro crítico: variável JWT_SECRET não definida no ambiente")
	}
	jwtSecret = []byte(secret)
}

// Claims define a estrutura interna do payload do token
type Claims struct {
	UserID  uint `json:"userId"` // camelCase para compatibilidade JSON
	IsAdmin bool `json:"isAdmin"`
	jwt.RegisteredClaims
}

// GerarToken cria um JWT válido por 24h
func GerarToken(userID uint, isAdmin bool) (string, error) {
	expiration := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		UserID:  userID,
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidarToken decodifica e valida o JWT
func ValidarToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("falha ao analisar token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token inválido")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("não foi possível extrair claims do token")
	}

	return claims, nil
}
