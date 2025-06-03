package negociacao

import (
	"encoding/json"
	"net/http"
	"strconv"
	"api/internal/notificacao"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type Handler struct {
	DB         *gorm.DB
	Repository Repository
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		DB:         db,
		Repository: NewRepository(),
	}
}
// POST /negociacoes
func (h *Handler) CriarNegociacao(w http.ResponseWriter, r *http.Request) {
	var n Negociacao
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}

	// Verifica se já existe negociação com o mesmo CNPJ
	if existente, _ := h.Repository.BuscarPorCNPJ(h.DB, n.CNPJ); existente != nil {
		notificacao.EnviarWebhookAlerta(n.CNPJ)
	}

	if err := h.Repository.Criar(h.DB, &n); err != nil {
		http.Error(w, "Erro ao criar negociação", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(n)
}



// GET /negociacoes/consultor/{id}
func (h *Handler) ListarPorConsultor(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	negociacoes, err := h.Repository.ListarPorConsultor(h.DB, uint(id))
	if err != nil {
		http.Error(w, "Erro ao buscar negociações", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(negociacoes)
}

// GET /negociacoes
func (h *Handler) ListarTodos(w http.ResponseWriter, r *http.Request) {
	negociacoes, err := h.Repository.ListarTodos(h.DB)
	if err != nil {
		http.Error(w, "Erro ao listar negociações", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(negociacoes)
}

// GET /negociacoes/{id}
func (h *Handler) BuscarPorID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	n, err := h.Repository.BuscarPorID(h.DB, uint(id))
	if err != nil {
		http.Error(w, "Negociação não encontrada", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(n)
}

// PUT /negociacoes/{id}
func (h *Handler) Atualizar(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	var dados Negociacao
	if err := json.NewDecoder(r.Body).Decode(&dados); err != nil {
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}

	if err := h.Repository.Atualizar(h.DB, uint(id), &dados); err != nil {
		http.Error(w, "Erro ao atualizar negociação", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Negociação atualizada com sucesso"))
}

// DELETE /negociacoes/{id}
func (h *Handler) Deletar(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	if err := h.Repository.Deletar(h.DB, uint(id)); err != nil {
		http.Error(w, "Erro ao excluir negociação", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Negociação excluída com sucesso"))
}
