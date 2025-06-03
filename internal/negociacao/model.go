package negociacao

import (
	"api/internal/comentario"
	"api/internal/contrato"
	"gorm.io/gorm"
)

type Negociacao struct {
	gorm.Model
	Nome                string                  `json:"nome"`
	Sobrenome           string                  `json:"sobrenome"`
	Telefone            string                  `json:"telefone"`
	CNPJ                string                  `json:"cnpj"`
	AnexoFatura         string                  `json:"anexoFatura"`
	AnexoEstudo         string                  `json:"anexoEstudo"`
	ContratoKC          string                  `json:"contratoKC"`
	AnexoContratoSocial string                  `json:"anexoContratoSocial"`
	Status              string                  `json:"status"`
	Produtos            string                  `json:"produtos"`
	KromaTake           bool                    `json:"kromaTake"`
	UF                  string                  `json:"uf"`
	ConsultorID         uint                    `json:"consultorId"`

	Contrato    contrato.Contrato              `gorm:"foreignKey:NegociacaoID" json:"contrato"`
	Comentarios []comentario.Comentario        `gorm:"foreignKey:NegociacaoID" json:"comentarios"`
}
