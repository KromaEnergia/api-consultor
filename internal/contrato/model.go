package contrato

import (
	"time"

	"gorm.io/gorm"
)

// Contrato representa tanto contratos de gestão quanto de energia
// e carrega toda a lógica de pagamento (fee, unipay, mensal).
type Contrato struct {
	gorm.Model

	NegociacaoID uint `gorm:"not null;index" json:"negociacaoId"`
	ConsultorID  uint `gorm:"not null;index" json:"consultorId"`

	Tipo           string    `gorm:"size:50;not null" json:"tipo"` // "Gestão" ou "Energia"
	URL            string    `gorm:"not null"       json:"url"`    // link/payload do contrato
	DataAssinatura time.Time `json:"dataAssinatura"`               // quando foi assinado

	Valor            float64   `gorm:"not null"       json:"valor"`  // valor total do contrato
	InicioSuprimento time.Time `json:"inicioSuprimento"`             // primeira data de contagem
	FimSuprimento    time.Time `json:"fimSuprimento"`                // última data de contagem
	ValorIntegral    bool      `json:"valorIntegral"`                // se true, usa Valor inteiro; se false, paga por período
	Status           string    `gorm:"size:50"        json:"status"` // ex: "Em Andamento", "Concluído"...

	// Nova lógica de fee
	Fee        bool    `json:"fee"`        // habilita fee
	FeePercent float64 `json:"feePercent"` // percentual de fee (0-1)

	// Pagamento único (upfront)
	UniPay        bool    `json:"unipay"`        // habilita pagamento único
	UniPayPercent float64 `json:"unipayPercent"` // percentual para pagamento único (0-1)

	// Pagamento mensal
	MonPay   bool    `json:"monPay"`   // habilita pagamento mensal
	MonthPay float64 `json:"monthPay"` // valor mensal fixo
}
