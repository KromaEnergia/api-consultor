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
	return &Handler{
		DB:         db,
		Repository: NewRepository(),
	}
}

// POST /contratos
func (h *Handler) CriarContrato(w http.ResponseWriter, r *http.Request) {
	var c Contrato
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}
	if err := h.Repository.Criar(h.DB, &c); err != nil {
		http.Error(w, "Erro ao criar contrato", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

// GET /contratos/negociacao/{id}
func (h *Handler) BuscarPorNegociacao(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	c, err := h.Repository.BuscarPorNegociacao(h.DB, uint(id))
	if err != nil {
		http.Error(w, "Contrato não encontrado", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(c)
}
