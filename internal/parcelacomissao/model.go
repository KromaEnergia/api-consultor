package parcelacomissao

import (
	"time"

	"gorm.io/gorm"
)

// ParcelaComissao representa uma única parcela da comissão de gestão mensal.
type ParcelaComissao struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	CalculoComissaoID uint       `gorm:"not null;index" json:"calculoComissaoId"` // Chave para a tabela pai
	Valor             float64    `gorm:"not null;default:0" json:"valor"`
	DataVencimento    time.Time  `gorm:"not null" json:"dataVencimento"`
	Status            string     `gorm:"size:50;not null;default:'Pendente';index" json:"status"` // "Pendente", "Pago", "Atrasado"
	DataPagamento     *time.Time `json:"dataPagamento"`                                           // Ponteiro para aceitar nulo
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

// Migrate cria a tabela no banco de dados.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&ParcelaComissao{})
}
