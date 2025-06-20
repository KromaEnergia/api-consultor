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
	Status           string    `json:"negociacao.Status"`

	// Nova lógica de fee
	Fee        bool    `json:"fee"`        // habilita fee
	FeePercent float64 `json:"feePercent"` // percentual de fee (0-1)

	// Pagamento único
	UniPay        bool    `json:"unipay"`        // habilita pagamento único
	UniPayPercent float64 `json:"unipayPercent"` // percentual para pagamento único (0-1)

	// Pagamento mensal
	MonPay   bool    `json:"monPay"`   // habilita pagamento mensal
	MonthPay float64 `json:"monthPay"` // valor mensal
}
