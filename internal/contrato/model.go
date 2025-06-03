package contrato

import (
	"time"
	"gorm.io/gorm"
)

type Contrato struct {
	gorm.Model
	NegociacaoID    uint      `json:"negociacaoId"`                   // 🔗 FK para Negociacao
	Valor           float64   `json:"valor"`                          // 💰 Valor do contrato
	InicioSuprimento time.Time `json:"inicioSuprimento"`              // 📅 Data início
	FimSuprimento    time.Time `json:"fimSuprimento"`                 // 📅 Data fim
	ValorIntegral   bool      `json:"valorIntegral"`                  // ✅ Se o valor da comissão foi pago integralmente no início
}
