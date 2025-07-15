package calculocomissao

import (
	"time"

	"github.com/KromaEnergia/api-consultor/internal/parcelacomissao"
	"gorm.io/gorm"
)

// CalculoComissao representa o cálculo de comissão vinculado a uma negociação
type CalculoComissao struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	NegociacaoID          uint      `gorm:"not null;index" json:"negociacaoId"`
	Status                string    `gorm:"size:100;not null;default:'Pendente';index" json:"status"`
	ModalidadeRecebimento string    `gorm:"size:255;not null" json:"modalidadeRecebimento"`
	Fee                   float64   `gorm:"not null;default:0" json:"fee"`
	Volume                float64   `gorm:"not null;default:0" json:"volume"`
	ValorGestaoMensal     float64   `gorm:"not null;default:0" json:"valorGestaoMensal"`
	EnergiaMensal         float64   `gorm:"not null;default:0" json:"energiaMensal"`
	PossuiComissaoGestao  bool      `gorm:"not null;default:false" json:"possuiComissaoGestao"`
	TotalReceber          float64   `gorm:"not null;default:0" json:"totalReceber"`
	InicioContrato        time.Time `json:"inicioContrato"`
	TerminioContrato      time.Time `json:"terminioContrato"`
	ModoPagamento         string    `json:"modoPagamento"`
	// Configuração de geração automática de parcelas
	QtdParcelas int       `gorm:"not null;default:0" json:"qtdParcelas"`
	DataGeracao time.Time `gorm:"not null" json:"dataGeracao"`

	// Associação com as parcelas geradas
	Parcelas []parcelacomissao.ParcelaComissao `gorm:"foreignKey:CalculoComissaoID;constraint:OnDelete:CASCADE" json:"parcelas"`

	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

// Migrate cria a tabela no banco de dados e aplica relacionamentos
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&CalculoComissao{})
}
