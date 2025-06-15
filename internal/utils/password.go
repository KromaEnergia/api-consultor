package utils

import "golang.org/x/crypto/bcrypt"

// HashSenha retorna o hash bcrypt da senha em texto
func HashSenha(senha string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	return string(hash), err
}

// CheckSenha compara hash bcrypt com a senha em texto e retorna true se bater
func CheckSenha(hash, senha string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(senha))
	return err == nil
}
