package consultor

import (
	"github.com/KromaEnergia/api-consultor/internal/contrato"
	"github.com/KromaEnergia/api-consultor/internal/negociacao"

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
	IsAdmin               bool                    `json:"isAdmin"` // ← nova flag
	Negociacoes           []negociacao.Negociacao `gorm:"foreignKey:ConsultorID" json:"negociacoes"`
	Contratos             []contrato.Contrato     `gorm:"foreignKey:ConsultorID" json:"contratos"`
}
