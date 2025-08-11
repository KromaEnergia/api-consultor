package models

import "gorm.io/gorm"

// Comentario representa um comentário ou um registro de histórico em uma negociação.
type Comentario struct {
	gorm.Model
	Texto        string `json:"texto"`        // Texto do comentário.
	NegociacaoID uint   `json:"negociacaoId"` // Chave estrangeira para a negociação.
	ConsultorID  uint   `json:"consultorId"`  // Chave estrangeira para o consultor autor. 0 se for do sistema.

	// Novo campo para identificar explicitamente comentários do sistema.
	System bool `gorm:"default:false" json:"system"`

	// A relação abaixo pode causar ciclos de importação se Negociacao também importar Comentario.
	// Considere removê-la se não for estritamente necessária na serialização JSON.
	// Negociacao    Negociacao `gorm:"foreignKey:NegociacaoID" json:"negociacao"`
}
