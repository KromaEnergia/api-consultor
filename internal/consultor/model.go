package consultor

import (
	"api/internal/contrato" // ← importe aqui
	"api/internal/negociacao"

	"gorm.io/gorm"
)

type Consultor struct {
	gorm.Model
	Nome                  string                  `json:"nome"`
	Sobrenome             string                  `json:"sobrenome"`
	CNPJ                  string                  `json:"cnpj" gorm:"unique"`
	Email                 string                  `json:"email" gorm:"unique"`
	Telefone              string                  `json:"telefone"`
	Foto                  string                  `json:"foto"`
	Senha                 string                  `json:"-"`
	PrecisaRedefinirSenha bool                    `json:"-"`
	Negociacoes           []negociacao.Negociacao `gorm:"foreignKey:ConsultorID" json:"negociacoes"`
	Contratos             []contrato.Contrato     `gorm:"foreignKey:ConsultorID" json:"contratos"` // ← adicione isto
}
