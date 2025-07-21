// internal/negociacao/handler.go
package negociacao

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/KromaEnergia/api-consultor/internal/auth"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// AdicionarArquivosRequest representa o payload para adicionar arquivos
type AdicionarArquivosRequest struct {
	NovosArquivos []string `json:"novosArquivos"`
}

// Handler encapsula DB e repository
type Handler struct {
	DB         *gorm.DB
	Repository Repository
}

// NewHandler cria um novo handler de negociações
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		DB:         db,
		Repository: NewRepository(),
	}
}

// negociacaoDTO é o payload de criação/atualização (sem ConsultorID)
type negociacaoDTO struct {
	Nome                string   `json:"nome"`
	Contato             string   `json:"contato"`
	NumeroDoContato     string   `json:"numeroDoContato"`
	Telefone            string   `json:"telefone"`
	CNPJ                string   `json:"cnpj"`
	Logo                string   `json:"logo"`
	AnexoFatura         string   `json:"anexoFatura"`
	AnexoEstudo         string   `json:"anexoEstudo"`
	ContratoKC          string   `json:"contratoKC"`
	AnexoContratoSocial string   `json:"anexoContratoSocial"`
	Status              string   `json:"status"`
	Produtos            []string `json:"produtos"`
	KromaTake           bool     `json:"kromaTake"`
	UF                  string   `json:"uf"`
	Arquivos            []string `json:"arquivos"`
}

// Criar trata POST /negociacoes
func (h *Handler) Criar(w http.ResponseWriter, r *http.Request) {
	userVal := r.Context().Value(auth.UsuarioIDKey)
	if userVal == nil {
		http.Error(w, "não autenticado", http.StatusUnauthorized)
		return
	}
	consultorID := userVal.(uint)

	var dto negociacaoDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	n := Negociacao{
		Nome:                dto.Nome,
		Contato:             dto.Contato,
		NumeroDoContato:     dto.NumeroDoContato,
		Telefone:            dto.Telefone,
		CNPJ:                dto.CNPJ,
		Logo:                dto.Logo,
		AnexoFatura:         dto.AnexoFatura,
		AnexoEstudo:         dto.AnexoEstudo,
		ContratoKC:          dto.ContratoKC,
		AnexoContratoSocial: dto.AnexoContratoSocial,
		Status:              dto.Status,
		KromaTake:           dto.KromaTake,
		UF:                  dto.UF,
		ConsultorID:         consultorID,
		Arquivos:            dto.Arquivos,
	}

	if err := h.Repository.Salvar(h.DB, &n); err != nil {
		http.Error(w, "Erro ao salvar negociação", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(n)
}

// ListarPorConsultor trata GET /consultores/{id}/negociacoes
func (h *Handler) ListarPorConsultor(w http.ResponseWriter, r *http.Request) {
	cid, _ := strconv.Atoi(mux.Vars(r)["id"])
	list, err := h.Repository.ListarPorConsultor(h.DB, uint(cid))
	if err != nil {
		http.Error(w, "Erro ao listar negociações", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(list)
}

// BuscarPorID trata GET /negociacoes/{id}
func (h *Handler) BuscarPorID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	n, err := h.Repository.BuscarPorID(h.DB, uint(id))
	if err != nil {
		http.Error(w, "Negociação não encontrada", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(n)
}

// Atualizar trata PUT /negociacoes/{id}
func (h *Handler) Atualizar(w http.ResponseWriter, r *http.Request) {
	idParam := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	userVal := r.Context().Value(auth.UsuarioIDKey)
	if userVal == nil {
		http.Error(w, "não autenticado", http.StatusUnauthorized)
		return
	}
	consultorID := userVal.(uint)

	var existing Negociacao
	if err := h.DB.First(&existing, id).Error; err != nil {
		http.Error(w, "Negociação não encontrada", http.StatusNotFound)
		return
	}

	var dto negociacaoDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	existing.Nome = dto.Nome
	existing.Contato = dto.Contato
	existing.NumeroDoContato = dto.NumeroDoContato
	existing.Telefone = dto.Telefone
	existing.CNPJ = dto.CNPJ
	existing.Logo = dto.Logo
	existing.AnexoFatura = dto.AnexoFatura
	existing.AnexoEstudo = dto.AnexoEstudo
	existing.ContratoKC = dto.ContratoKC
	existing.AnexoContratoSocial = dto.AnexoContratoSocial
	existing.Status = dto.Status
	existing.KromaTake = dto.KromaTake
	existing.UF = dto.UF
	existing.ConsultorID = consultorID
	existing.Arquivos = dto.Arquivos

	if err := h.Repository.Atualizar(h.DB, &existing); err != nil {
		http.Error(w, "Erro ao atualizar negociação", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existing)
}

// Deletar trata DELETE /negociacoes/{id}
func (h *Handler) Deletar(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	if err := h.Repository.Deletar(h.DB, uint(id)); err != nil {
		http.Error(w, "Erro ao excluir negociação", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// AdicionarArquivos adiciona novas URLs ao slice 'Arquivos'
func (h *Handler) AdicionarArquivos(w http.ResponseWriter, r *http.Request) {
	idParam := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	userVal := r.Context().Value(auth.UsuarioIDKey)
	if userVal == nil {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}
	consultorID := userVal.(uint)

	var req AdicionarArquivosRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if len(req.NovosArquivos) == 0 {
		http.Error(w, "A lista 'novosArquivos' não pode estar vazia", http.StatusBadRequest)
		return
	}

	var existente Negociacao
	if err := h.DB.First(&existente, id).Error; err != nil {
		http.Error(w, "Negociação não encontrada", http.StatusNotFound)
		return
	}

	isAdmin := r.Context().Value(auth.IsAdminKey).(bool)
	if !isAdmin && existente.ConsultorID != consultorID {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	existente.Arquivos = append(existente.Arquivos, req.NovosArquivos...)
	if err := h.Repository.Atualizar(h.DB, &existente); err != nil {
		http.Error(w, "Erro ao salvar os novos arquivos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(existente)
}

// RemoverProduto trata DELETE /negociacoes/{id}/produtos/{idx}
func (h *Handler) RemoverProduto(w http.ResponseWriter, r *http.Request) {
	idParam := mux.Vars(r)["id"]
	idxParam := mux.Vars(r)["idx"]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	idx, err := strconv.Atoi(idxParam)
	if err != nil {
		http.Error(w, "Índice inválido", http.StatusBadRequest)
		return
	}

	userVal := r.Context().Value(auth.UsuarioIDKey)
	if userVal == nil {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}
	consultorID := userVal.(uint)

	var existente Negociacao
	if err := h.DB.First(&existente, id).Error; err != nil {
		http.Error(w, "Negociação não encontrada", http.StatusNotFound)
		return
	}

	isAdmin := r.Context().Value(auth.IsAdminKey).(bool)
	if !isAdmin && existente.ConsultorID != consultorID {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	if idx < 0 || idx >= len(existente.Produtos) {
		http.Error(w, "Índice de produto inválido", http.StatusBadRequest)
		return
	}

	existente.Produtos = append(existente.Produtos[:idx], existente.Produtos[idx+1:]...)
	if err := h.Repository.Atualizar(h.DB, &existente); err != nil {
		http.Error(w, "Erro ao remover produto", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existente)
}

// RemoverArquivo trata DELETE /negociacoes/{id}/arquivos/{idx}
func (h *Handler) RemoverArquivo(w http.ResponseWriter, r *http.Request) {
	idParam := mux.Vars(r)["id"]
	idxParam := mux.Vars(r)["idx"]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	idx, err := strconv.Atoi(idxParam)
	if err != nil {
		http.Error(w, "Índice inválido", http.StatusBadRequest)
		return
	}

	userVal := r.Context().Value(auth.UsuarioIDKey)
	if userVal == nil {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}
	consultorID := userVal.(uint)

	var existente Negociacao
	if err := h.DB.First(&existente, id).Error; err != nil {
		http.Error(w, "Negociação não encontrada", http.StatusNotFound)
		return
	}

	isAdmin := r.Context().Value(auth.IsAdminKey).(bool)
	if !isAdmin && existente.ConsultorID != consultorID {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	if idx < 0 || idx >= len(existente.Arquivos) {
		http.Error(w, "Índice de arquivo inválido", http.StatusBadRequest)
		return
	}

	existente.Arquivos = append(existente.Arquivos[:idx], existente.Arquivos[idx+1:]...)
	if err := h.Repository.Atualizar(h.DB, &existente); err != nil {
		http.Error(w, "Erro ao remover arquivo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existente)
}

/*
Rotas sugeridas:
router.HandleFunc("/negociacoes/{id}/produtos/{idx}", handler.RemoverProduto).Methods("DELETE")
router.HandleFunc("/negociacoes/{id}/arquivos/{idx}", handler.RemoverArquivo).Methods("DELETE")
*/
