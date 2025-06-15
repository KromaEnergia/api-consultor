// internal/negociacao/model.go
package negociacao

import (
	"time"

	"github.com/KromaEnergia/api-consultor/internal/comentario"
	"github.com/KromaEnergia/api-consultor/internal/contrato"
	"gorm.io/gorm"
)

// Negociacao representa uma oportunidade de negócio de um consultor
type Negociacao struct {
	ID        uint           `gorm:"primaryKey" json:"negociacaoId"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`

	Nome                string `json:"nome"`
	Contato             string `json:"contato"`
	Telefone            string `json:"telefone"`
	CNPJ                string `json:"cnpj"`
	Logo                string `json:"logo"`
	AnexoFatura         string `json:"anexoFatura"`
	AnexoEstudo         string `json:"anexoEstudo"`
	ContratoKC          string `json:"contratoKC"`
	AnexoContratoSocial string `json:"anexoContratoSocial"`
	Status              string `json:"status"`
	Produtos            string `json:"produtos"`
	KromaTake           bool   `json:"kromaTake"`
	UF                  string `json:"uf"`
	ConsultorID         uint   `json:"consultorId"`

	// Relação 1-1 com Contrato
	Contrato contrato.Contrato `gorm:"foreignKey:NegociacaoID" json:"contrato"`
	// Relação 1-N com Comentarios
	Comentarios []comentario.Comentario `gorm:"foreignKey:NegociacaoID" json:"comentarios"`
}
