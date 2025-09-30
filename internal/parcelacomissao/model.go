// internal/parcelacomissao/model.go
package parcelacomissao

import (
	"time"

	"gorm.io/gorm"
)

// ParcelaComissao representa uma única parcela da comissão de gestão mensal.
type ParcelaComissao struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	CalculoComissaoID uint       `gorm:"not null;index" json:"calculoComissaoId"`
	Valor             float64    `gorm:"not null;default:0" json:"valor"`
	VolumeMensal      float64    `gorm:"not null;default:0" json:"volumeMensal"`
	Anexo             string     `gorm:"size:255" json:"anexo"`
	NotaFiscal        string     `gorm:"size:255" json:"notaFiscal"`
	DataVencimento    time.Time  `gorm:"not null" json:"dataVencimento"`
	Status            string     `gorm:"size:50;not null;default:'Pendente';index" json:"status"`
	DataPagamento     *time.Time `json:"dataPagamento"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

// Migrate cria a tabela no banco de dados.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&ParcelaComissao{})
}
