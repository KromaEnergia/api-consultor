// internal/calculocomissao/dto.go
package calculocomissao

type CreateCalculoDTO struct {
	ModalidadeRecebimento string  `json:"modalidadeRecebimento"`
	Fee                   float64 `json:"fee"`
	Volume                float64 `json:"volume"`
	ValorGestaoMensal     float64 `json:"valorGestaoMensal"`
	EnergiaMensal         float64 `json:"energiaMensal"`
	PossuiComissaoGestao  bool    `json:"possuiComissaoGestao"`
	TotalReceber          float64 `json:"totalReceber"`

	ModoPagamento                 string  `json:"modoPagamento"`
	ValorPagamentoInicial         float64 `json:"valorPagamentoInicial"`
	DataPagamentoInicial          string  `json:"dataPagamentoInicial"`
	ValorPagamentoFinal           float64 `json:"valorPagamentoFinal"`
	DataPagamentoFinal            string  `json:"dataPagamentoFinal"`
	ValorPrimeiraParcela          float64 `json:"valorPrimeiraParcela"`
	DataVencimentoPrimeiraParcela string  `json:"dataVencimentoPrimeiraParcela"`
	ValorParcelaMensal            float64 `json:"valorParcelaMensal"`
	DataInicioParcelas            string  `json:"dataInicioParcelas"`
	QtdParcelas                   int     `json:"qtdParcelas"`

	InicioContrato   string `json:"inicioContrato"`
	TerminioContrato string `json:"terminioContrato"`
	DataGeracao      string `json:"dataGeracao"`
}
