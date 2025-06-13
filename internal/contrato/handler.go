package contrato

import (
	"encoding/json"
	"net/http"
	"strconv"

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

// POST /negociacoes/{id}/contrato
func (h *Handler) CriarParaNegociacao(w http.ResponseWriter, r *http.Request) {
	negID, _ := strconv.Atoi(mux.Vars(r)["id"])
	var c Contrato
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	c.NegociacaoID = uint(negID)
	if err := h.Repository.Salvar(h.DB, &c); err != nil {
		http.Error(w, "Erro ao salvar contrato", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

// GET /negociacoes/{id}/contrato
func (h *Handler) BuscarPorNegociacao(w http.ResponseWriter, r *http.Request) {
	negID, _ := strconv.Atoi(mux.Vars(r)["id"])
	c, err := h.Repository.BuscarPorNegociacao(h.DB, uint(negID))
	if err != nil {
		http.Error(w, "Contrato não encontrado", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(c)
}

// GET /consultores/{id}/contratos
func (h *Handler) ListarPorConsultor(w http.ResponseWriter, r *http.Request) {
	consID, _ := strconv.Atoi(mux.Vars(r)["id"])
	list, err := h.Repository.ListarPorConsultor(h.DB, uint(consID))
	if err != nil {
		http.Error(w, "Erro ao listar contratos", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(list)
}

// PUT /contratos/{id}
func (h *Handler) Atualizar(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	var c Contrato
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	c.ID = uint(id)
	if err := h.Repository.Atualizar(h.DB, &c); err != nil {
		http.Error(w, "Erro ao atualizar contrato", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(c)
}

// DELETE /contratos/{id}
func (h *Handler) Deletar(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	if err := h.Repository.Deletar(h.DB, uint(id)); err != nil {
		http.Error(w, "Erro ao excluir contrato", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
