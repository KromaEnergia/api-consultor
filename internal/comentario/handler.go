package comentario

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

func (h *Handler) CriarComentario(w http.ResponseWriter, r *http.Request) {
	var c Comentario
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}

	if err := h.Repository.Criar(h.DB, &c); err != nil {
		http.Error(w, "Erro ao criar comentário", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

func (h *Handler) ListarPorNegociacao(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	comentarios, err := h.Repository.ListarPorNegociacao(h.DB, uint(id))
	if err != nil {
		http.Error(w, "Erro ao listar comentários", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(comentarios)
}

func (h *Handler) RemoverComentario(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	if err := h.Repository.Remover(h.DB, uint(id)); err != nil {
		http.Error(w, "Erro ao remover comentário", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Comentário removido com sucesso"))
}
// GET /comentarios
func (h *Handler) ListarTodos(w http.ResponseWriter, r *http.Request) {
	comentarios, err := h.Repository.ListarTodos(h.DB)
	if err != nil {
		http.Error(w, "Erro ao listar comentários", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(comentarios)
}

// GET /comentarios/{id}
func (h *Handler) BuscarPorID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	comentario, err := h.Repository.BuscarPorID(h.DB, uint(id))
	if err != nil {
		http.Error(w, "Comentário não encontrado", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(comentario)
}

// PUT /comentarios/{id}
func (h *Handler) Atualizar(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	var payload struct {
		Texto string `json:"texto"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}

	if err := h.Repository.Atualizar(h.DB, uint(id), payload.Texto); err != nil {
		http.Error(w, "Erro ao atualizar comentário", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Comentário atualizado com sucesso"))
}
