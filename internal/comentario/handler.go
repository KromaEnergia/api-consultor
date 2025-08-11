package comentario

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/KromaEnergia/api-consultor/internal/auth" // Importe o pacote de autenticação
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

// CriarComentario trata da requisição POST /negociacoes/{id}/comentarios
func (h *Handler) CriarComentario(w http.ResponseWriter, r *http.Request) {
	// 1. Extrair o ID da negociação da URL
	idStr := mux.Vars(r)["id"]
	negID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID de negociação inválido", http.StatusBadRequest)
		return
	}

	// 2. Decodificar o corpo da requisição
	var req CriarComentarioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if req.Texto == "" {
		http.Error(w, "O campo 'texto' é obrigatório", http.StatusBadRequest)
		return
	}

	var consultorID uint
	isSystem := req.IsSystemComment

	// 3. Determinar o autor e o tipo de comentário
	if isSystem {
		// Para comentários do sistema, o ID do consultor é 0.
		consultorID = 0
	} else {
		// Para comentários de usuários, pegamos o ID do contexto de autenticação.
		userVal := r.Context().Value(auth.UsuarioIDKey)
		if userVal == nil {
			http.Error(w, "Não autenticado", http.StatusUnauthorized)
			return
		}
		consultorID = userVal.(uint)
	}

	// 4. Criar a entidade Comentario com o novo campo 'System'
	c := models.Comentario{
		Texto:        req.Texto,
		NegociacaoID: uint(negID),
		ConsultorID:  consultorID, // Será 0 se for do sistema
		System:       isSystem,    // Define se o comentário é do sistema
	}

	// 5. Salvar usando o repositório
	if err := h.Repository.Criar(h.DB, &c); err != nil {
		http.Error(w, "Erro ao criar comentário", http.StatusInternalServerError)
		return
	}

	// 6. Retornar o comentário criado
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

// ListarPorNegociacao trata GET /negociacoes/{id}/comentarios
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

// RemoverComentario trata DELETE /comentarios/{id}
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

// ListarTodos trata GET /comentarios
func (h *Handler) ListarTodos(w http.ResponseWriter, r *http.Request) {
	comentarios, err := h.Repository.ListarTodos(h.DB)
	if err != nil {
		http.Error(w, "Erro ao listar comentários", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(comentarios)
}

// BuscarPorID trata GET /comentarios/{id}
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

// Atualizar trata PUT /comentarios/{id}
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
