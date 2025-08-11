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

type ParcelaUpdateDTO struct {
	Valor          float64   `json:"valor"`
	DataVencimento time.Time `json:"dataVencimento"`
	Status         string    `json:"status"`
	Anexo          string    `json:"anexo"`
	VolumeMensal   float64   `json:"volumeMensal"`
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

// Adicione este código ao seu arquivo de handler (ex: parcelacomissao/handler.go)

// UpdateAnexo trata POST /parcelas/{pid}/anexo
// Usado para adicionar ou atualizar o link de um anexo.
func (h *Handler) UpdateAnexo(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(mux.Vars(r)["pid"])
	if err != nil {
		http.Error(w, "ID da parcela inválido", http.StatusBadRequest)
		return
	}

	var payload struct {
		Anexo string `json:"anexo"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON mal formado", http.StatusBadRequest)
		return
	}

	// Você pode adicionar validações aqui, como verificar se é uma URL válida.

	// Chama o método do repositório para atualizar o anexo
	err = h.Repo.UpdateAnexo(uint(pid), payload.Anexo)
	if err != nil {
		http.Error(w, "Erro ao atualizar anexo da parcela", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK) // Responde com 200 OK sem corpo
	w.Write([]byte(`{"message": "Anexo atualizado com sucesso"}`))
}

// DeleteAnexo trata DELETE /parcelas/{pid}/anexo
// Usado para remover o link de um anexo.
func (h *Handler) DeleteAnexo(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(mux.Vars(r)["pid"])
	if err != nil {
		http.Error(w, "ID da parcela inválido", http.StatusBadRequest)
		return
	}

	// Chama o método do repositório para remover o anexo (definindo-o como "")
	err = h.Repo.UpdateAnexo(uint(pid), "")
	if err != nil {
		http.Error(w, "Erro ao remover anexo da parcela", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK) // Responde com 200 OK sem corpo
	w.Write([]byte(`{"message": "Anexo removido com sucesso"}`))
}

// NOVO MÉTODO: Update trata PUT /parcelas/{pid}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(mux.Vars(r)["pid"])
	if err != nil {
		http.Error(w, "ID da parcela inválido", http.StatusBadRequest)
		return
	}

	// Primeiro, busca a parcela existente para garantir que ela existe.
	// (Isto assume que seu repositório tem um método FindByID)
	parcelaExistente, err := h.Repo.FindByID(uint(pid))
	if err != nil {
		http.Error(w, "Parcela não encontrada", http.StatusNotFound)
		return
	}

	// Decodifica o corpo da requisição no nosso DTO.
	var payload ParcelaUpdateDTO
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON mal formado", http.StatusBadRequest)
		return
	}

	// Atualiza os campos da parcela existente com os dados do payload.
	parcelaExistente.Valor = payload.Valor
	parcelaExistente.DataVencimento = payload.DataVencimento
	parcelaExistente.Status = payload.Status
	parcelaExistente.Anexo = payload.Anexo
	parcelaExistente.VolumeMensal = payload.VolumeMensal

	// Se o status for alterado para "Pago" e ainda não houver data de pagamento,
	// registra a data de pagamento atual.
	if payload.Status == "Pago" && parcelaExistente.DataPagamento == nil {
		now := time.Now()
		parcelaExistente.DataPagamento = &now
	}

	// Salva a parcela atualizada no banco.
	if err := h.Repo.Update(parcelaExistente); err != nil {
		http.Error(w, "Erro ao atualizar a parcela", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(parcelaExistente)
}
