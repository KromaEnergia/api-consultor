// internal/consultor/utils.go
package consultor

import "golang.org/x/crypto/bcrypt"

// CheckSenha compara o hash armazenado com a senha em texto
func CheckSenha(hash, senha string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(senha))
	return err == nil
}
