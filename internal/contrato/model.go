package contrato

import (
	"time"

	"gorm.io/gorm"
)

type Contrato struct {
	gorm.Model
	NegociacaoID     uint      `json:"negociacaoId"`
	ConsultorID      uint      `json:"consultorId"`
	Valor            float64   `json:"valor"`
	InicioSuprimento time.Time `json:"inicioSuprimento"`
	FimSuprimento    time.Time `json:"fimSuprimento"`
	ValorIntegral    bool      `json:"valorIntegral"`
}
