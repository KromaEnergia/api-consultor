// internal/produtos/model.go
package produtos

import "gorm.io/gorm"

type Produto struct {
	ID           uint    `gorm:"primaryKey" json:"id"`
	Tipo         string  `gorm:"size:255;not null" json:"tipo"`
	Ativo        bool    `gorm:"not null;default:false" json:"ativo"`
	Comissao     float64 `gorm:"not null;default:0" json:"comissao"`
	Fee          float64 `gorm:"not null;default:0" json:"fee"`
	NegociacaoID uint    `gorm:"not null;index" json:"negociacao_id"`
}

// AutoMigrate em algum init:
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&Produto{})
}
