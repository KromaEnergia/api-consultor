package models

import "gorm.io/gorm"

// Comentario representa um comentário ou um registro de histórico em uma negociação.
type Comentario struct {
	gorm.Model
	Texto        string `json:"texto"`                     // Texto do comentário
	NegociacaoID uint   `json:"negociacaoId" gorm:"index"` // FK da negociação

	// Autor CONSULTOR (0 se não for consultor; para sistema/admin)
	ConsultorID uint `json:"consultorId" gorm:"index"`

	// Autor COMERCIAL/Admin (nulo quando não for comercial)
	ComercialID *uint `json:"comercialId" gorm:"index"`

	// Mantém a coluna existente "system" no banco, mas no código chamamos de IsSystem.
	IsSystem bool `gorm:"column:system;default:false" json:"system"`

	// Opcional: flag semântica; pode ser derivada de (ComercialID != nil)
	IsAdminAuthor bool `gorm:"column:is_admin_author;default:false" json:"isAdminAuthor"`
}
