// models/negociacao.go
package models

import (
	"time"

	"github.com/KromaEnergia/api-consultor/internal/calculocomissao"
	"github.com/KromaEnergia/api-consultor/internal/contrato"
	produto "github.com/KromaEnergia/api-consultor/internal/produtos"
	"gorm.io/gorm"
)

// Convenção de status textual para anexos
const (
	StatusPendente = "Pendente"
	StatusEnviado  = "Enviado"
	StatusValidado = "Validado"
)

// MultiAnexo representa anexos múltiplos com status textual (para Fatura).
type MultiAnexo struct {
	Itens  []string `json:"itens"`            // 1+ links
	Status string   `json:"status,omitempty"` // "Pendente" | "Enviado" | "Validado" (ou outros que você aceitar)
}

type Negociacao struct {
	ID        uint           `gorm:"primaryKey" json:"negociacaoId"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`

	// Dados básicos
	Nome            string `json:"nome"`
	Email           string `json:"email"`
	Contato         string `json:"contato"`
	NumeroDoContato string `json:"numeroDoContato"`
	Telefone        string `json:"telefone"`
	CNPJ            string `json:"cnpj"`

	// ---- Anexos simples + status textual ----
	Logo                      string `json:"logo"`
	LogoStatus                string `json:"logoStatus"` // ex.: "Pendente", "Enviado", "Validado"
	AnexoEstudo               string `json:"anexoEstudo"`
	AnexoEstudoStatus         string `json:"anexoEstudoStatus"`
	ContratoKC                string `json:"contratoKC"`
	ContratoKCStatus          string `json:"contratoKCStatus"`
	AnexoContratoSocial       string `json:"anexoContratoSocial"`
	AnexoContratoSocialStatus string `json:"anexoContratoSocialStatus"`

	// ---- Novos anexos solicitados + status textual ----
	AnexoProcuracao                string `json:"anexoProcuracao"`
	AnexoProcuracaoStatus          string `json:"anexoProcuracaoStatus"`
	AnexoRepresentanteLegal        string `json:"anexoRepresentanteLegal"`
	AnexoRepresentanteLegalStatus  string `json:"anexoRepresentanteLegalStatus"`
	AnexoEstudoDeViabilidade       string `json:"anexoEstudoDeViabilidade"`
	AnexoEstudoDeViabilidadeStatus string `json:"anexoEstudoDeViabilidadeStatus"`

	// ---- Fatura: objeto (vários links) + status textual ----
	AnexoFatura MultiAnexo `gorm:"type:jsonb;serializer:json" json:"anexoFatura"`

	// Múltiplos anexos “soltos”
	Arquivos []string `gorm:"type:jsonb;serializer:json" json:"arquivos"`

	// Outros campos de negócio
	Status      string              `json:"status"` // status geral da negociação
	Contratos   []contrato.Contrato `gorm:"foreignKey:NegociacaoID;constraint:OnDelete:CASCADE" json:"contratos"`
	Produtos    []produto.Produto   `gorm:"foreignKey:NegociacaoID" json:"produtos"`
	KromaTake   bool                `json:"kromaTake"`
	UF          string              `json:"uf"`
	ConsultorID uint                `json:"consultorId"`

	Comentarios      []Comentario                      `gorm:"foreignKey:NegociacaoID" json:"comentarios"`
	CalculosComissao []calculocomissao.CalculoComissao `gorm:"foreignKey:NegociacaoID;constraint:OnDelete:CASCADE" json:"calculosComissao"`
}
