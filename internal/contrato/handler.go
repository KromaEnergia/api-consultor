package contrato

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type Handler struct {
	DB         *gorm.DB
	Repository Repository
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{DB: db, Repository: NewRepository()}
}

// DTO para criação/atualização
type contratoDTO struct {
	Tipo             string    `json:"tipo"`
	URL              string    `json:"url"`
	DataAssinatura   time.Time `json:"dataAssinatura"`
	Valor            float64   `json:"valor"`
	InicioSuprimento time.Time `json:"inicioSuprimento"`
	FimSuprimento    time.Time `json:"fimSuprimento"`
	ValorIntegral    bool      `json:"valorIntegral"`
	Status           string    `json:"status"`

	Fee        bool    `json:"fee"`
	FeePercent float64 `json:"feePercent"`

	UniPay        bool    `json:"unipay"`
	UniPayPercent float64 `json:"unipayPercent"`

	MonPay   bool    `json:"monPay"`
	MonthPay float64 `json:"monthPay"`
}

// CriarParaNegociacao insere um novo contrato
func (h *Handler) CriarParaNegociacao(w http.ResponseWriter, r *http.Request) {
	// 1. Extrai ID da negociação da URL
	negID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID de negociação inválido", http.StatusBadRequest)
		return
	}

	// 2. Busca consultorID associado à negociação
	var neg struct{ ConsultorID uint }
	if err := h.DB.Table("negociacaos").
		Select("consultor_id").
		Where("id = ?", negID).
		Scan(&neg).Error; err != nil {
		http.Error(w, "Negociação não encontrada", http.StatusNotFound)
		return
	}

	// 3. Decodifica JSON para DTO
	var dto contratoDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// 4. Mapeia DTO para model
	c := Contrato{
		NegociacaoID:     uint(negID),
		ConsultorID:      neg.ConsultorID,
		Tipo:             dto.Tipo,
		URL:              dto.URL,
		DataAssinatura:   dto.DataAssinatura,
		Valor:            dto.Valor,
		InicioSuprimento: dto.InicioSuprimento,
		FimSuprimento:    dto.FimSuprimento,
		ValorIntegral:    dto.ValorIntegral,
		Status:           dto.Status,

		Fee:        dto.Fee,
		FeePercent: dto.FeePercent,

		UniPay:        dto.UniPay,
		UniPayPercent: dto.UniPayPercent,

		MonPay:   dto.MonPay,
		MonthPay: dto.MonthPay,
	}

	// 5. Persiste no banco
	if err := h.Repository.Salvar(h.DB, &c); err != nil {
		http.Error(w, "Erro ao salvar contrato", http.StatusInternalServerError)
		return
	}

	// 6. Retorna o contrato criado
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

// BuscarPorNegociacao retorna o contrato
func (h *Handler) BuscarPorNegociacao(w http.ResponseWriter, r *http.Request) {
	negID, _ := strconv.Atoi(mux.Vars(r)["id"])
	c, err := h.Repository.BuscarPorNegociacao(h.DB, uint(negID))
	if err != nil {
		http.Error(w, "Contrato não encontrado", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(c)
}

// ListarPorConsultor retorna contratos de um consultor
func (h *Handler) ListarPorConsultor(w http.ResponseWriter, r *http.Request) {
	consID, _ := strconv.Atoi(mux.Vars(r)["id"])
	list, err := h.Repository.ListarPorConsultor(h.DB, uint(consID))
	if err != nil {
		http.Error(w, "Erro ao listar contratos", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(list)
}

// Atualizar modifica um contrato existente
func (h *Handler) Atualizar(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var existing Contrato
	if err := h.DB.First(&existing, id).Error; err != nil {
		http.Error(w, "Contrato não encontrado", http.StatusNotFound)
		return
	}

	var dto contratoDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	existing.Valor = dto.Valor
	existing.InicioSuprimento = dto.InicioSuprimento
	existing.FimSuprimento = dto.FimSuprimento
	existing.ValorIntegral = dto.ValorIntegral
	existing.Status = dto.Status
	existing.Fee = dto.Fee
	existing.FeePercent = dto.FeePercent
	existing.UniPay = dto.UniPay
	existing.UniPayPercent = dto.UniPayPercent
	existing.MonPay = dto.MonPay
	existing.MonthPay = dto.MonthPay

	if err := h.Repository.Atualizar(h.DB, &existing); err != nil {
		http.Error(w, "Erro ao atualizar contrato", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(existing)
}

// Deletar remove um contrato
func (h *Handler) Deletar(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	if err := h.Repository.Deletar(h.DB, uint(id)); err != nil {
		http.Error(w, "Erro ao excluir contrato", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
