package negociacao

import (
	"time"

	"github.com/KromaEnergia/api-consultor/internal/calculocomissao"
	"github.com/KromaEnergia/api-consultor/internal/comentario"
	"github.com/KromaEnergia/api-consultor/internal/contrato"

	// "github.com/KromaEnergia/api-consultor/internal/contrato"
	produto "github.com/KromaEnergia/api-consultor/internal/produtos"
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

	// Suporta múltiplos anexos de arquivos em JSONB
	Arquivos []string `gorm:"type:jsonb;serializer:json" json:"arquivos"`

	Status    string              `json:"status"`
	Contratos []contrato.Contrato `gorm:"foreignKey:NegociacaoID;constraint:OnDelete:CASCADE" json:"contratos"`
	// --- Aqui entra Produtos ---
	// Suporta múltiplos produtos em JSONB via relação 1-N
	Produtos []produto.Produto `gorm:"foreignKey:NegociacaoID" json:"produtos"`

	KromaTake   bool   `json:"kromaTake"`
	UF          string `json:"uf"`
	ConsultorID uint   `json:"consultorId"`

	// Relação 1-1 com Contrato

	// Relação 1-N com Comentarios
	Comentarios []comentario.Comentario `gorm:"foreignKey:NegociacaoID" json:"comentarios"`

	CalculosComissao []calculocomissao.CalculoComissao `gorm:"foreignKey:NegociacaoID;constraint:OnDelete:CASCADE" json:"calculosComissao"`
}
