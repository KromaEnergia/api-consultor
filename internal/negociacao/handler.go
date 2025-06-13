package negociacao

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

// POST /negociacoes
func (h *Handler) Criar(w http.ResponseWriter, r *http.Request) {
	var n Negociacao
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if err := h.Repository.Salvar(h.DB, &n); err != nil {
		http.Error(w, "Erro ao salvar negociação", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(n)
}

// GET /consultores/{id}/negociacoes
func (h *Handler) ListarPorConsultor(w http.ResponseWriter, r *http.Request) {
	consultorID, _ := strconv.Atoi(mux.Vars(r)["id"])
	list, err := h.Repository.ListarPorConsultor(h.DB, uint(consultorID))
	if err != nil {
		http.Error(w, "Erro ao listar negociações", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(list)
}

// GET /negociacoes/{id}
func (h *Handler) BuscarPorID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	n, err := h.Repository.BuscarPorID(h.DB, uint(id))
	if err != nil {
		http.Error(w, "Negociação não encontrada", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(n)
}

// PUT /negociacoes/{id}
func (h *Handler) Atualizar(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	var n Negociacao
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	n.ID = uint(id)
	if err := h.Repository.Atualizar(h.DB, &n); err != nil {
		http.Error(w, "Erro ao atualizar", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(n)
}

// DELETE /negociacoes/{id}
func (h *Handler) Deletar(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	if err := h.Repository.Deletar(h.DB, uint(id)); err != nil {
		http.Error(w, "Erro ao excluir", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
