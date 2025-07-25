package calculocomissao

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/KromaEnergia/api-consultor/internal/parcelacomissao"
	"github.com/gorilla/mux"
)

// Handler gerencia rotas de cálculo de comissão
type Handler struct {
	Repo *Repository
}

// NewHandler cria um novo Handler
func NewHandler(repo *Repository) *Handler {
	return &Handler{Repo: repo}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	// 1) pega ID da negociação
	negID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID de negociação inválido", http.StatusBadRequest)
		return
	}

	// 2) decodifica no DTO
	var dto CreateCalculoDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "JSON mal formado", http.StatusBadRequest)
		return
	}

	// 3) parse de strings para time.Time
	parse := func(s string) time.Time {
		t, _ := time.Parse(time.RFC3339, s)
		return t
	}
	dataGeracao := parse(dto.DataGeracao)
	dataPagamentoInicial := parse(dto.DataPagamentoInicial)  // para dividirInicialEmDuas
	dataPagamentoFinal := parse(dto.DataPagamentoFinal)      // para dividirInicialEmDuas
	dataPrimeira := parse(dto.DataVencimentoPrimeiraParcela) // para pagamentoInicialEParcelas
	dataInicio := parse(dto.DataInicioParcelas)              // para pagamentoInicialEParcelas e parcelasIguais
	inicioContrato := parse(dto.InicioContrato)
	terminoContrato := parse(dto.TerminioContrato)

	// 4) monta o model
	calc := CalculoComissao{
		NegociacaoID:          uint(negID),
		ModalidadeRecebimento: dto.ModalidadeRecebimento,
		Fee:                   dto.Fee,
		Volume:                dto.Volume,
		ValorGestaoMensal:     dto.ValorGestaoMensal,
		EnergiaMensal:         dto.EnergiaMensal,
		PossuiComissaoGestao:  dto.PossuiComissaoGestao,
		TotalReceber:          dto.TotalReceber,
		InicioContrato:        inicioContrato,
		TerminioContrato:      terminoContrato,
		QtdParcelas:           dto.QtdParcelas,
		DataGeracao:           dataGeracao,
	}

	// 5) persiste o cálculo
	if err := h.Repo.Create(&calc); err != nil {
		http.Error(w, "Erro ao criar cálculo", http.StatusInternalServerError)
		return
	}

	// 6) gera parcelas conforme o modo de pagamento
	parcRepo := parcelacomissao.NewRepository(h.Repo.DB)
	var parcelas []*parcelacomissao.ParcelaComissao

	switch dto.ModoPagamento {
	case "dividirInicialEmDuas":
		// duas cobranças avulsas
		parcelas = append(parcelas,
			&parcelacomissao.ParcelaComissao{
				CalculoComissaoID: calc.ID,
				Valor:             dto.ValorPagamentoInicial,
				DataVencimento:    dataPagamentoInicial,
				Status:            "Pendente",
			},
			&parcelacomissao.ParcelaComissao{
				CalculoComissaoID: calc.ID,
				Valor:             dto.ValorPagamentoFinal,
				DataVencimento:    dataPagamentoFinal,
				Status:            "Pendente",
			},
		)

	case "pagamentoInicialEParcelas":
		// primeira parcela “maior”
		if dto.ValorPrimeiraParcela > 0 {
			parcelas = append(parcelas, &parcelacomissao.ParcelaComissao{
				CalculoComissaoID: calc.ID,
				Valor:             dto.ValorPrimeiraParcela,
				DataVencimento:    dataPrimeira,
				Status:            "Pendente",
			})
		}
		// demais parcelas iguais
		for i := 0; i < dto.QtdParcelas; i++ {
			parcelas = append(parcelas, &parcelacomissao.ParcelaComissao{
				CalculoComissaoID: calc.ID,
				Valor:             dto.ValorParcelaMensal,
				DataVencimento:    dataInicio.AddDate(0, i, 0),
				Status:            "Pendente",
			})
		}

	case "parcelasIguais":
		// todas as parcelas iguais, sem pagamento inicial
		for i := 0; i < dto.QtdParcelas; i++ {
			parcelas = append(parcelas, &parcelacomissao.ParcelaComissao{
				CalculoComissaoID: calc.ID,
				Valor:             dto.ValorParcelaMensal,
				DataVencimento:    dataInicio.AddDate(0, i, 0),
				Status:            "Pendente",
			})
		}

	default:
		http.Error(w, "Modo de pagamento inválido", http.StatusBadRequest)
		return
	}

	if err := parcRepo.CreateInBatch(parcelas); err != nil {
		http.Error(w, "Erro ao criar parcelas", http.StatusInternalServerError)
		return
	}

	// 7) recarrega e retorna
	h.Repo.DB.Preload("Parcelas").First(&calc, calc.ID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(calc)
}

// List trata GET /negociacoes/{id}/calculos-comissao
// Aceita um query param opcional `status` para filtrar os resultados.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	negID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID de negociação inválido", http.StatusBadRequest)
		return
	}

	status := r.URL.Query().Get("status")

	var list []CalculoComissao
	if status != "" {
		list, err = h.Repo.FindByNegociacaoAndStatus(uint(negID), status)
	} else {
		list, err = h.Repo.FindByNegociacao(uint(negID))
	}
	if err != nil {
		http.Error(w, "Erro ao buscar cálculos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

// Get trata GET /negociacoes/{id}/calculos-comissao/{cid}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	cid, err := strconv.Atoi(mux.Vars(r)["cid"])
	if err != nil {
		http.Error(w, "ID do cálculo inválido", http.StatusBadRequest)
		return
	}

	calc, err := h.Repo.FindByID(uint(cid))
	if err != nil {
		http.Error(w, "Cálculo não encontrado", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(calc)
}

// Update trata PUT /negociacoes/{id}/calculos-comissao/{cid}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	cid, err := strconv.Atoi(mux.Vars(r)["cid"])
	if err != nil {
		http.Error(w, "ID do cálculo inválido", http.StatusBadRequest)
		return
	}

	calc, err := h.Repo.FindByID(uint(cid))
	if err != nil {
		http.Error(w, "Cálculo não encontrado", http.StatusNotFound)
		return
	}

	var payload CalculoComissao
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON mal formado", http.StatusBadRequest)
		return
	}

	// Atualiza campos permitidos
	calc.Status = payload.Status
	calc.ModalidadeRecebimento = payload.ModalidadeRecebimento
	calc.Fee = payload.Fee
	calc.InicioContrato = payload.InicioContrato
	calc.TerminioContrato = payload.TerminioContrato
	calc.Volume = payload.Volume
	calc.PossuiComissaoGestao = payload.PossuiComissaoGestao
	calc.TotalReceber = payload.TotalReceber
	calc.ValorGestaoMensal = payload.ValorGestaoMensal
	calc.EnergiaMensal = payload.EnergiaMensal

	if err := h.Repo.Update(calc); err != nil {
		http.Error(w, "Erro ao atualizar cálculo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(calc)
}

// UpdateStatus trata PATCH /negociacoes/{id}/calculos-comissao/{cid}/status
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	cid, err := strconv.Atoi(mux.Vars(r)["cid"])
	if err != nil {
		http.Error(w, "ID do cálculo inválido", http.StatusBadRequest)
		return
	}

	var payload struct {
		Status       string  `json:"status"`
		TotalReceber float64 `json:"totalReceber"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON mal formado ou campos inválidos", http.StatusBadRequest)
		return
	}

	if payload.Status == "" {
		http.Error(w, "O campo 'status' é obrigatório", http.StatusBadRequest)
		return
	}

	err = h.Repo.UpdateStatusAndTotal(uint(cid), payload.Status, payload.TotalReceber)
	if err != nil {
		http.Error(w, "Erro ao atualizar status do cálculo", http.StatusInternalServerError)
		return
	}

	calc, err := h.Repo.FindByID(uint(cid))
	if err != nil {
		http.Error(w, "Cálculo não encontrado após atualização", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(calc)
}

// Delete trata DELETE /negociacoes/{id}/calculos-comissao/{cid}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	cid, err := strconv.Atoi(mux.Vars(r)["cid"])
	if err != nil {
		http.Error(w, "ID do cálculo inválido", http.StatusBadRequest)
		return
	}

	calc, err := h.Repo.FindByID(uint(cid))
	if err != nil {
		http.Error(w, "Cálculo não encontrado", http.StatusNotFound)
		return
	}

	if err := h.Repo.Delete(calc); err != nil {
		http.Error(w, "Erro ao deletar cálculo", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
