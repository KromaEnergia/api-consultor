// internal/comercial/model.go
package comercial

import (
	"time"

	"github.com/KromaEnergia/api-consultor/internal/consultor"
)

type Comercial struct {
	ID          uint                  `gorm:"primaryKey" json:"id"`
	Nome        string                `gorm:"size:100;not null" json:"nome"`
	Sobrenome   string                `gorm:"size:100;not null" json:"sobrenome"`
	Documento   string                `gorm:"size:20;not null" json:"documento"`
	Email       string                `gorm:"size:100;uniqueIndex;not null" json:"email"`
	Password    string                `gorm:"size:255;not null" json:"-"` // não expõe a senha no JSON
	Telefone    string                `gorm:"size:20" json:"telefone"`
	Foto        string                `gorm:"size:255" json:"foto"`
	IsAdmin     bool                  `gorm:"default:false" json:"isAdmin"`
	Consultores []consultor.Consultor `gorm:"foreignKey:ComercialID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"consultores"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}
