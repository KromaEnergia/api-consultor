package consultor

import (
	"api/internal/negociacao"
	"gorm.io/gorm"
		"api/internal/contrato"
	"api/internal/negociacao"
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
	Negociacoes           []negociacao.Negociacao `gorm:"foreignKey:ConsultorID"`
}
