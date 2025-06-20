package contrato

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/KromaEnergia/api-consultor/internal/auth"
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
	Valor            float64   `json:"valor"`
	InicioSuprimento time.Time `json:"inicioSuprimento"`
	FimSuprimento    time.Time `json:"fimSuprimento"`
	ValorIntegral    bool      `json:"valorIntegral"`
	Status           string    `json:"status"`
	Fee              bool      `json:"fee"`
	FeePercent       float64   `json:"feePercent"`
	UniPay           bool      `json:"unipay"`
	UniPayPercent    float64   `json:"unipayPercent"`
	MonPay           bool      `json:"monPay"`
	MonthPay         float64   `json:"monthPay"`
}

// CriarParaNegociacao insere um novo contrato
func (h *Handler) CriarParaNegociacao(w http.ResponseWriter, r *http.Request) {
	negID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID de negociação inválido", http.StatusBadRequest)
		return
	}

	var neg struct{ ConsultorID uint }
	if err := h.DB.Table("negociacaos").
		Select("consultor_id").
		Where("id = ?", negID).
		Scan(&neg).Error; err != nil {
		http.Error(w, "Negociação não encontrada", http.StatusNotFound)
		return
	}

	var dto contratoDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	c := Contrato{
		NegociacaoID:     uint(negID),
		ConsultorID:      neg.ConsultorID,
		Valor:            dto.Valor,
		InicioSuprimento: dto.InicioSuprimento,
		FimSuprimento:    dto.FimSuprimento,
		ValorIntegral:    dto.ValorIntegral,
		Status:           dto.Status,
		Fee:              dto.Fee,
		FeePercent:       dto.FeePercent,
		UniPay:           dto.UniPay,
		UniPayPercent:    dto.UniPayPercent,
		MonPay:           dto.MonPay,
		MonthPay:         dto.MonthPay,
	}

	if err := h.Repository.Salvar(h.DB, &c); err != nil {
		http.Error(w, "Erro ao salvar contrato", http.StatusInternalServerError)
		return
	}

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

func (h *Handler) Comissoes(w http.ResponseWriter, r *http.Request) {
	type ComissaoTotal struct {
		ComissoesRecebidas float64 `json:"comissoesRecebidas"`
		ComissoesAReceber  float64 `json:"comissoesAReceber"`
	}

	// 1) Pega o ID do consultor autenticado
	userIDVal := r.Context().Value(auth.UsuarioIDKey)
	if userIDVal == nil {
		http.Error(w, "não autenticado", http.StatusUnauthorized)
		return
	}
	consultorID := userIDVal.(uint)

	// 2) Busca apenas contratos deste consultor
	//    **e** cujas negociações correspondentes tenham status = "Contrato Fechado"
	var contratos []Contrato
	if err := h.DB.
		Joins("JOIN negociacaos n ON n.id = contratos.negociacao_id").
		Where("contratos.consultor_id = ? AND n.status = ?", consultorID, "Contrato Fechado").
		Find(&contratos).Error; err != nil {
		http.Error(w, "Erro ao buscar contratos", http.StatusInternalServerError)
		return
	}

	// 3) Se não vier nada, logo não há comissões a somar
	if len(contratos) == 0 {
		resp := ComissaoTotal{ComissoesRecebidas: 0, ComissoesAReceber: 0}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	now := time.Now()
	var totalRec, totalARec float64

	// função auxiliar para meses inclusivos
	monthsBetween := func(start, end time.Time) int {
		y1, m1, _ := start.Date()
		y2, m2, _ := end.Date()
		diff := int((y2-y1)*12 + int(m2-m1))
		return diff + 1
	}

	for _, c := range contratos {
		// FEE: dividido em 2 parcelas, início e fim
		if c.Fee {
			feeTotal := c.Valor * c.FeePercent
			half := feeTotal / 2
			// primeira metade no mês de início
			if now.After(c.InicioSuprimento) || now.Equal(c.InicioSuprimento) {
				totalRec += half
			} else {
				totalARec += half
			}
			// segunda metade no mês de fim
			if now.After(c.FimSuprimento) || now.Equal(c.FimSuprimento) {
				totalRec += half
			} else {
				totalARec += half
			}
		}

		// UniPay (único pagamento)
		if c.UniPay {
			val := c.Valor * c.UniPayPercent
			if now.After(c.InicioSuprimento) || now.Equal(c.InicioSuprimento) {
				totalRec += val
			} else {
				totalARec += val
			}
		}

		// MonPay (mensal)
		if c.MonPay {
			months := monthsBetween(c.InicioSuprimento, c.FimSuprimento)
			if months <= 0 {
				months = 1
			}
			totalValue := c.Valor

			switch {
			case now.Before(c.InicioSuprimento):
				totalARec += totalValue
			case now.After(c.FimSuprimento) || now.Equal(c.FimSuprimento):
				totalRec += totalValue
			default:
				elapsed := monthsBetween(c.InicioSuprimento, now)
				if elapsed > months {
					elapsed = months
				}
				monthly := totalValue / float64(months)
				rec := monthly * float64(elapsed)
				totalRec += rec
				totalARec += totalValue - rec
			}
		}
	}

	resp := ComissaoTotal{
		ComissoesRecebidas: totalRec,
		ComissoesAReceber:  totalARec,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
