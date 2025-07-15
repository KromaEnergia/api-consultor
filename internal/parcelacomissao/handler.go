package parcelacomissao

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type Handler struct {
	Repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{Repo: repo}
}

// List trata GET /calculos-comissao/{cid}/parcelas
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	cid, err := strconv.Atoi(mux.Vars(r)["cid"])
	if err != nil {
		http.Error(w, "ID do cálculo de comissão inválido", http.StatusBadRequest)
		return
	}

	parcelas, err := h.Repo.ListByCalculoID(uint(cid))
	if err != nil {
		http.Error(w, "Erro ao buscar parcelas", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(parcelas)
}

// UpdateStatus trata PATCH /parcelas/{pid}
// Usado para marcar uma parcela como "Paga".
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(mux.Vars(r)["pid"])
	if err != nil {
		http.Error(w, "ID da parcela inválido", http.StatusBadRequest)
		return
	}

	var payload struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON mal formado", http.StatusBadRequest)
		return
	}

	if payload.Status != "Pago" { // Exemplo de regra de negócio
		http.Error(w, "Ação permitida apenas para marcar como 'Pago'", http.StatusBadRequest)
		return
	}

	// Atualiza o status e seta a data de pagamento para agora
	err = h.Repo.UpdateStatus(uint(pid), payload.Status, time.Now())
	if err != nil {
		http.Error(w, "Erro ao atualizar status da parcela", http.StatusInternalServerError)
		return
	}

	// Retorna a parcela atualizada
	parcela, err := h.Repo.FindByID(uint(pid))
	if err != nil {
		http.Error(w, "Erro ao buscar parcela atualizada", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(parcela)
}
