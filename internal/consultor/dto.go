package consultor

type ResumoConsultorDTO struct {
	ID                uint     `json:"id"`
	Nome              string   `json:"nome"`
	Sobrenome         string   `json:"sobrenome"`
	Email             string   `json:"email"`
	CNPJ              string   `json:"cnpj"`
	Telefone          string   `json:"telefone"`
	Foto              string   `json:"foto"`
	ContratosFechados int      `json:"contratosFechados"`
	NegociacoesAtivas int      `json:"negociacoesAtivas"`
	ComissaoRecebida  float64  `json:"comissaoRecebida"`
	ComissaoAReceber  float64  `json:"comissaoAReceber"`
	Produtos          []string `json:"produtos"` // lista de tipos de produto
}
