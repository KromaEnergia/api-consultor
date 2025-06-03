package auth

import (
	"encoding/json"
	"net/http"
	"yourapp/consultor"
	"yourapp/utils"

	"gorm.io/gorm"
)

func LoginHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			EmailOuCNPJ string `json:"emailOuCnpj"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		var user consultor.Consultor
		err := db.Where("email = ? OR cnpj = ?", req.EmailOuCNPJ, req.EmailOuCNPJ).First(&user).Error
		if err != nil {
			http.Error(w, "Usuário não encontrado", http.StatusUnauthorized)
			return
		}

		if user.PrecisaRedefinirSenha {
			// enviar senha temporária e pedir redefinição
			senhaTemporaria := utils.GerarSenhaTemporaria()
			user.Senha = utils.HashSenha(senhaTemporaria)
			db.Save(&user)
			// Enviar senha por email/SMS ou exibir no frontend
			w.Write([]byte("Senha temporária gerada. Redefina sua senha."))
			return
		}

		// autenticação normal aqui (não implementada agora)
	}
}
