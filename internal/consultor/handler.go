package consultor

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/KromaEnergia/api-consultor/internal/auth"
	"github.com/KromaEnergia/api-consultor/internal/contrato"
	"github.com/KromaEnergia/api-consultor/internal/negociacao"
	"github.com/KromaEnergia/api-consultor/internal/utils"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// request DTOs
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type createConsultorRequest struct {
	Nome      string `json:"nome"`
	Sobrenome string `json:"sobrenome"`
	CNPJ      string `json:"cnpj"`
	Email     string `json:"email"`
	Telefone  string `json:"telefone"`
	Foto      string `json:"foto"`
	Senha     string `json:"senha"`
	IsAdmin   bool   `json:"isAdmin"`
}

// Handler encapsula DB e repository
type Handler struct {
	DB         *gorm.DB
	Repository Repository
}

// NewHandler retorna um handler inicializado
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		DB:         db,
		Repository: NewRepository(),
	}
}

// Login gera um JWT para credenciais válidas
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "payload inválido", http.StatusBadRequest)
		return
	}

	// Busca usuário por email ou CNPJ
	user, err := h.Repository.BuscarPorEmailOuCNPJ(h.DB, req.Login)
	if err != nil {
		http.Error(w, "credenciais inválidas", http.StatusUnauthorized)
		return
	}

	// Verifica senha
	if !utils.CheckSenha(user.Senha, req.Password) {
		http.Error(w, "senha incorreta", http.StatusUnauthorized)
		return
	}

	// Gera token
	token, err := auth.GerarToken(user.ID, user.IsAdmin)
	if err != nil {
		http.Error(w, "erro ao gerar token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// CriarConsultor cadastra novo consultor (livre de autenticação)
func (h *Handler) CriarConsultor(w http.ResponseWriter, r *http.Request) {
	var req createConsultorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "payload inválido", http.StatusBadRequest)
		return
	}

	// Gera hash da senha
	hash, err := utils.HashSenha(req.Senha)
	if err != nil {
		http.Error(w, "erro ao processar senha", http.StatusInternalServerError)
		return
	}

	// Monta modelo
	c := Consultor{
		Nome:                  req.Nome,
		Sobrenome:             req.Sobrenome,
		CNPJ:                  req.CNPJ,
		Email:                 req.Email,
		Telefone:              req.Telefone,
		Foto:                  req.Foto,
		Senha:                 hash,
		PrecisaRedefinirSenha: false,
		IsAdmin:               req.IsAdmin,
	}

	if err := h.Repository.Salvar(h.DB, &c); err != nil {
		http.Error(w, "erro ao salvar consultor", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

// ListarConsultores retorna todos ou apenas o próprio registro
func (h *Handler) ListarConsultores(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UsuarioIDKey).(uint)
	isAdmin := r.Context().Value(auth.IsAdminKey).(bool)

	if isAdmin {
		consultores, err := h.Repository.ListarTodos(h.DB)
		if err != nil {
			http.Error(w, "erro ao listar consultores", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(consultores)
		return
	}

	// não-admin vê apenas o próprio
	obj, err := h.Repository.BuscarPorID(h.DB, userID)
	if err != nil {
		http.Error(w, "consultor não encontrado", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode([]Consultor{*obj})
}

// BuscarPorID retorna um consultor pelo ID
func (h *Handler) BuscarPorID(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UsuarioIDKey).(uint)
	isAdmin := r.Context().Value(auth.IsAdminKey).(bool)

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	if !isAdmin && uint(id) != userID {
		http.Error(w, "acesso negado", http.StatusForbidden)
		return
	}

	obj, err := h.Repository.BuscarPorID(h.DB, uint(id))
	if err != nil {
		http.Error(w, "consultor não encontrado", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(obj)
}

// AtualizarConsultor altera dados de um consultor existente
func (h *Handler) AtualizarConsultor(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UsuarioIDKey).(uint)
	isAdmin := r.Context().Value(auth.IsAdminKey).(bool)

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	if !isAdmin && uint(id) != userID {
		http.Error(w, "acesso negado", http.StatusForbidden)
		return
	}

	var dados Consultor
	if err := json.NewDecoder(r.Body).Decode(&dados); err != nil {
		http.Error(w, "payload inválido", http.StatusBadRequest)
		return
	}
	if err := h.Repository.Atualizar(h.DB, uint(id), &dados); err != nil {
		http.Error(w, "erro ao atualizar consultor", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("consultor atualizado com sucesso"))
}

// DeletarConsultor remove um consultor
func (h *Handler) DeletarConsultor(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UsuarioIDKey).(uint)
	isAdmin := r.Context().Value(auth.IsAdminKey).(bool)

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	if !isAdmin && uint(id) != userID {
		http.Error(w, "acesso negado", http.StatusForbidden)
		return
	}

	if err := h.Repository.Deletar(h.DB, uint(id)); err != nil {
		http.Error(w, "erro ao excluir consultor", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("consultor excluído com sucesso"))
}

// ObterResumoConsultor constrói e retorna o DTO de resumo
func (h *Handler) ObterResumoConsultor(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UsuarioIDKey).(uint)
	isAdmin := r.Context().Value(auth.IsAdminKey).(bool)

	idParam := userID
	if isAdmin {
		if idStr := mux.Vars(r)["id"]; idStr != "" {
			if i, err := strconv.Atoi(idStr); err == nil {
				idParam = uint(i)
			}
		}
	}

	consultorObj, err := h.Repository.BuscarPorID(h.DB, idParam)
	if err != nil {
		http.Error(w, "consultor não encontrado", http.StatusNotFound)
		return
	}

	negociacoes, _ := negociacao.NewRepository().ListarPorConsultor(h.DB, consultorObj.ID)
	contratos, _ := contrato.NewRepository().ListarPorConsultor(h.DB, consultorObj.ID)
	dto := MontarResumoConsultorDTO(*consultorObj, contratos, negociacoes)

	json.NewEncoder(w).Encode(dto)
}

// Me retorna o usuário logado
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UsuarioIDKey).(uint)

	var c Consultor
	if err := h.DB.First(&c, userID).Error; err != nil {
		http.Error(w, "consultor não encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}
