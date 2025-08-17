package comentario

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/KromaEnergia/api-consultor/internal/auth"
	"github.com/KromaEnergia/api-consultor/internal/models"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// Handler encapsula o DB e o Repository
type Handler struct {
	DB         *gorm.DB
	Repository Repository
}

// NewHandler cria um novo handler de comentários
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		DB:         db,
		Repository: NewRepository(),
	}
}

// CriarComentarioRequest define o corpo da requisição para criar um comentário.
type CriarComentarioRequest struct {
	Texto           string `json:"texto"`
	IsSystemComment bool   `json:"isSystemComment,omitempty"`
}

// helper para ponteiro
func ptr[T any](v T) *T { return &v }

// POST /negociacoes/{id}/comentarios
func (h *Handler) CriarComentario(w http.ResponseWriter, r *http.Request) {
	// 1) ID da negociação
	idStr := mux.Vars(r)["id"]
	negID, err := strconv.Atoi(idStr)
	if err != nil || negID <= 0 {
		http.Error(w, "ID de negociação inválido", http.StatusBadRequest)
		return
	}

	// 2) Body
	var req CriarComentarioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.Texto) == "" {
		http.Error(w, "O campo 'texto' é obrigatório", http.StatusBadRequest)
		return
	}

	// 3) Autenticação (claims no contexto)
	userVal := r.Context().Value(auth.CtxUserID)
	if userVal == nil {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}
	usuarioID, ok := userVal.(uint)
	if !ok || usuarioID == 0 {
		http.Error(w, "Falha ao obter usuário do contexto", http.StatusUnauthorized)
		return
	}
	isAdminVal := r.Context().Value(auth.CtxIsAdmin)
	isAdmin, _ := isAdminVal.(bool)

	// 4) Regra: comentário de sistema só admin
	if req.IsSystemComment && !isAdmin {
		http.Error(w, "Apenas admin pode criar comentário de sistema", http.StatusForbidden)
		return
	}

	// 5) Monta registro conforme autoria (NENHUMA mudança de model/DB)
	com := models.Comentario{
		NegociacaoID:  uint(negID),
		Texto:         strings.TrimSpace(req.Texto),
		IsSystem:      req.IsSystemComment,
		IsAdminAuthor: isAdmin, // opcional; pode ser derivado de ComercialID != nil
	}
	if req.IsSystemComment {
		// Comentário de sistema não tem autor consultor/comercial
		com.ConsultorID = 0
		com.ComercialID = nil
	} else if isAdmin {
		// Admin/Comercial vira "lado comercial"
		com.ConsultorID = 0
		com.ComercialID = ptr(usuarioID)
	} else {
		// Autor consultor
		com.ConsultorID = usuarioID
		com.ComercialID = nil
	}

	// 6) Persiste
	if err := h.DB.Create(&com).Error; err != nil {
		http.Error(w, "Erro ao salvar comentário", http.StatusInternalServerError)
		return
	}

	// 7) Resposta crua do model (sem DTO)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(com)
}

// GET /negociacoes/{id}/comentarios
func (h *Handler) ListarPorNegociacao(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	comentarios, err := h.Repository.ListarPorNegociacao(h.DB, uint(id))
	if err != nil {
		http.Error(w, "Erro ao listar comentários", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(comentarios)
}

// DELETE /comentarios/{id}
func (h *Handler) RemoverComentario(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	if err := h.Repository.Remover(h.DB, uint(id)); err != nil {
		http.Error(w, "Erro ao remover comentário", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Comentário removido com sucesso"))
}

// GET /comentarios
func (h *Handler) ListarTodos(w http.ResponseWriter, r *http.Request) {
	comentarios, err := h.Repository.ListarTodos(h.DB)
	if err != nil {
		http.Error(w, "Erro ao listar comentários", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(comentarios)
}

// GET /comentarios/{id}
func (h *Handler) BuscarPorID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	comentario, err := h.Repository.BuscarPorID(h.DB, uint(id))
	if err != nil {
		http.Error(w, "Comentário não encontrado", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(comentario)
}

// PUT /comentarios/{id}
func (h *Handler) Atualizar(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var payload struct {
		Texto string `json:"texto"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(payload.Texto) == "" {
		http.Error(w, "O campo 'texto' é obrigatório", http.StatusBadRequest)
		return
	}

	if err := h.Repository.Atualizar(h.DB, uint(id), strings.TrimSpace(payload.Texto)); err != nil {
		http.Error(w, "Erro ao atualizar comentário", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Comentário atualizado com sucesso"))
}
