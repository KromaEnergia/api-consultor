package consultor

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/KromaEnergia/api-consultor/internal/auth"
	"github.com/KromaEnergia/api-consultor/internal/contrato"
	"github.com/KromaEnergia/api-consultor/internal/negociacao"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// DTOs

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

// Handler encapsula DB e repo
type Handler struct {
	DB         *gorm.DB
	Repository Repository
}

// NewHandler cria Handler
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{DB: db, Repository: NewRepository()}
}

// Login gera JWT
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "payload inválido", http.StatusBadRequest)
		return
	}

	user, err := h.Repository.BuscarPorEmailOuCNPJ(h.DB, req.Login)
	if err != nil {
		http.Error(w, "credenciais inválidas", http.StatusUnauthorized)
		return
	}

	// Compare bcrypt hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Senha), []byte(req.Password)); err != nil {
		http.Error(w, "senha incorreta", http.StatusUnauthorized)
		return
	}

	token, err := auth.GerarToken(user.ID, user.IsAdmin)
	if err != nil {
		http.Error(w, "erro ao gerar token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

// CriarConsultor cadastro público
func (h *Handler) CriarConsultor(w http.ResponseWriter, r *http.Request) {
	var req createConsultorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "payload inválido", http.StatusBadRequest)
		return
	}

	// Hash bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Senha), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "erro ao processar senha", http.StatusInternalServerError)
		return
	}

	c := Consultor{
		Nome:                  req.Nome,
		Sobrenome:             req.Sobrenome,
		CNPJ:                  req.CNPJ,
		Email:                 req.Email,
		Telefone:              req.Telefone,
		Foto:                  req.Foto,
		Senha:                 string(hash),
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

// GET /consultores/me
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UsuarioIDKey).(uint)

	var c Consultor
	// Preload das associações negociacoes e contratos
	if err := h.DB.
		Preload("Negociacoes").
		Preload("Contratos").
		First(&c, userID).Error; err != nil {
		http.Error(w, "Consultor não encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}
