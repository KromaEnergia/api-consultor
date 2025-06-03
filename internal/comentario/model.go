package comentario

import "gorm.io/gorm"

type Comentario struct {
	gorm.Model
	Texto        string `json:"texto"`
	NegociacaoID uint   `json:"negociacaoId"`
}
