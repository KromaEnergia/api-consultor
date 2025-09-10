package comercial

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/KromaEnergia/api-consultor/internal/auth"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
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

// POST /comerciais/login
// Valida email/senha, emite access token RS256 e seta refresh token em cookie httpOnly.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "payload inválido", http.StatusBadRequest)
		return
	}

	user, err := h.Repository.FindByEmail(h.DB, req.Email)
	if err != nil {
		http.Error(w, "credenciais inválidas", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "senha incorreta", http.StatusUnauthorized)
		return
	}

	// Emite access token e seta refresh (httpOnly) no cookie
	access, err := auth.IssueTokensOnLogin(h.DB, w, user.ID /* isAdmin */, user.IsAdmin)
	if err != nil {
		fmt.Print("Erro ao gerar tokens: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access_token": access,
		"token_type":   "Bearer",
		"expires_in":   int(auth.AccessTTL.Seconds()),
	})
}

// POST /comerciais
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateComercialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "payload inválido", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Senha), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "erro ao processar senha", http.StatusInternalServerError)
		return
	}

	c := Comercial{
		Nome:      req.Nome,
		Sobrenome: req.Sobrenome,
		Documento: req.Documento,
		Email:     req.Email,
		Telefone:  req.Telefone,
		Foto:      req.Foto,
		Password:  string(hash),
		IsAdmin:   req.IsAdmin,
	}

	if err := h.Repository.Save(h.DB, &c); err != nil {
		http.Error(w, "erro ao salvar comercial", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(c)
}

// GET /comerciais
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	// Se esta rota já estiver atrás de auth.RequireAdmin no router,
	// este check é redundante — mas mantém como fallback.
	isAdmin, _ := r.Context().Value(auth.CtxIsAdmin).(bool)
	if !isAdmin {
		http.Error(w, "acesso negado", http.StatusForbidden)
		return
	}

	list, err := h.Repository.ListAll(h.DB)
	if err != nil {
		http.Error(w, "erro ao listar comerciais", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

// GET /comerciais/{id}
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	isAdmin, _ := r.Context().Value(auth.CtxIsAdmin).(bool)
	userID, _ := r.Context().Value(auth.CtxUserID).(uint)
	if !isAdmin && uint(id) != userID {
		http.Error(w, "acesso negado", http.StatusForbidden)
		return
	}

	obj, err := h.Repository.FindByID(h.DB, uint(id))
	if err != nil {
		http.Error(w, "comercial não encontrado", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(obj)
}

// PUT /comerciais/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	isAdmin, _ := r.Context().Value(auth.CtxIsAdmin).(bool)
	userID, _ := r.Context().Value(auth.CtxUserID).(uint)
	if !isAdmin && uint(id) != userID {
		http.Error(w, "acesso negado", http.StatusForbidden)
		return
	}

	var req UpdateComercialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "payload inválido", http.StatusBadRequest)
		return
	}

	if err := h.Repository.Update(h.DB, uint(id), &req); err != nil {
		http.Error(w, "erro ao atualizar comercial", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("comercial atualizado com sucesso"))
}

// DELETE /comerciais/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	isAdmin, _ := r.Context().Value(auth.CtxIsAdmin).(bool)
	userID, _ := r.Context().Value(auth.CtxUserID).(uint)
	if !isAdmin && uint(id) != userID {
		http.Error(w, "acesso negado", http.StatusForbidden)
		return
	}

	if err := h.Repository.Delete(h.DB, uint(id)); err != nil {
		http.Error(w, "erro ao excluir comercial", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("comercial excluído com sucesso"))
}

// GET /comerciais/me
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(auth.CtxUserID).(uint)

	var c Comercial
	// Carrega também o slice de Consultores e suas relações (Negociações, Contratos)
	if err := h.DB.
		Preload("Consultores", func(db *gorm.DB) *gorm.DB {
			return db.
				Preload("Negociacoes").
				Preload("Contratos").
				Preload("Negociacoes.Contratos").
				Preload("Negociacoes.Produtos").
				Preload("Negociacoes.CalculosComissao").
				Preload("Negociacoes.CalculosComissao.Parcelas")
		}).
		First(&c, userID).Error; err != nil {
		http.Error(w, "Comercial não encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(c)
}
