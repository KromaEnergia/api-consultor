// internal/negociacao/handler.go
package negociacao

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/KromaEnergia/api-consultor/internal/auth"
	"github.com/KromaEnergia/api-consultor/internal/models"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// AdicionarArquivosRequest representa o payload para adicionar arquivos
type AdicionarArquivosRequest struct {
	NovosArquivos []string `json:"novosArquivos"`
}

// PatchAnexoEstudoRequest define o corpo da requisição para atualizar o anexo do estudo.
type PatchAnexoEstudoRequest struct {
	AnexoEstudo string `json:"anexoEstudo"`
}

// PatchContratoKCRequest define o corpo da requisição para atualizar o contrato KC.
type PatchContratoKCRequest struct {
	ContratoKC string `json:"contratoKC"`
}

// Handler encapsula DB e repository
type Handler struct {
	DB         *gorm.DB
	Repository Repository
}

// CORREÇÃO: Struct para o payload de atualização de status definida corretamente.
type atualizarStatusRequest struct {
	Status string `json:"status"`
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
	Email               string   `json:"email"`
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
	userVal := r.Context().Value(auth.CtxUserID)
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

	n := models.Negociacao{
		Nome:                dto.Nome,
		Email:               dto.Email,
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
	_ = json.NewEncoder(w).Encode(n)
}

// ListarPorConsultor trata GET /consultores/{id}/negociacoes
func (h *Handler) ListarPorConsultor(w http.ResponseWriter, r *http.Request) {
	cid, _ := strconv.Atoi(mux.Vars(r)["id"])

	isAdmin, _ := r.Context().Value(auth.CtxIsAdmin).(bool)
	userID, _ := r.Context().Value(auth.CtxUserID).(uint)
	if !isAdmin && uint(cid) != userID {
		http.Error(w, "acesso negado", http.StatusForbidden)
		return
	}

	list, err := h.Repository.ListarPorConsultor(h.DB, uint(cid))
	if err != nil {
		http.Error(w, "Erro ao listar negociações", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(list)
}

// BuscarPorID trata GET /negociacoes/{id}
func (h *Handler) BuscarPorID(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	n, err := h.Repository.BuscarPorID(h.DB, uint(id))
	if err != nil {
		http.Error(w, "Negociação não encontrada", http.StatusNotFound)
		return
	}

	// Permissão: admin ou dono da negociação
	isAdmin, _ := r.Context().Value(auth.CtxIsAdmin).(bool)
	userID, _ := r.Context().Value(auth.CtxUserID).(uint)
	if !isAdmin && n.ConsultorID != userID {
		http.Error(w, "acesso negado", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(n)
}

// Atualizar trata PUT /negociacoes/{id}
func (h *Handler) Atualizar(w http.ResponseWriter, r *http.Request) {
	idParam := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	userVal := r.Context().Value(auth.CtxUserID)
	if userVal == nil {
		http.Error(w, "não autenticado", http.StatusUnauthorized)
		return
	}
	consultorID := userVal.(uint)
	isAdmin, _ := r.Context().Value(auth.CtxIsAdmin).(bool)

	var existing models.Negociacao
	if err := h.DB.First(&existing, id).Error; err != nil {
		http.Error(w, "Negociação não encontrada", http.StatusNotFound)
		return
	}

	// Permissão: admin ou dono
	if !isAdmin && existing.ConsultorID != consultorID {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	var dto negociacaoDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	existing.Nome = dto.Nome
	existing.Contato = dto.Contato
	existing.Email = dto.Email
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
	_ = json.NewEncoder(w).Encode(existing)
}

// Deletar trata DELETE /negociacoes/{id}
func (h *Handler) Deletar(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	// Permissão: admin ou dono
	var existente models.Negociacao
	if err := h.DB.First(&existente, id).Error; err != nil {
		http.Error(w, "Negociação não encontrada", http.StatusNotFound)
		return
	}
	isAdmin, _ := r.Context().Value(auth.CtxIsAdmin).(bool)
	userID, _ := r.Context().Value(auth.CtxUserID).(uint)
	if !isAdmin && existente.ConsultorID != userID {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

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

	userVal := r.Context().Value(auth.CtxUserID)
	if userVal == nil {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}
	consultorID := userVal.(uint)
	isAdmin, _ := r.Context().Value(auth.CtxIsAdmin).(bool)

	var req AdicionarArquivosRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if len(req.NovosArquivos) == 0 {
		http.Error(w, "A lista 'novosArquivos' não pode estar vazia", http.StatusBadRequest)
		return
	}

	var existente models.Negociacao
	if err := h.DB.First(&existente, id).Error; err != nil {
		http.Error(w, "Negociação não encontrada", http.StatusNotFound)
		return
	}

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
	_ = json.NewEncoder(w).Encode(existente)
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

	userVal := r.Context().Value(auth.CtxUserID)
	if userVal == nil {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}
	consultorID := userVal.(uint)
	isAdmin, _ := r.Context().Value(auth.CtxIsAdmin).(bool)

	var existente models.Negociacao
	if err := h.DB.First(&existente, id).Error; err != nil {
		http.Error(w, "Negociação não encontrada", http.StatusNotFound)
		return
	}

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
	_ = json.NewEncoder(w).Encode(existente)
}

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

	userVal := r.Context().Value(auth.CtxUserID)
	if userVal == nil {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}
	consultorID := userVal.(uint)
	isAdmin, _ := r.Context().Value(auth.CtxIsAdmin).(bool)

	var existente models.Negociacao
	if err := h.DB.First(&existente, id).Error; err != nil {
		http.Error(w, "Negociação não encontrada", http.StatusNotFound)
		return
	}

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
	_ = json.NewEncoder(w).Encode(existente)
}

// PatchAnexoEstudo trata PATCH /negociacoes/{id}/anexo-estudo
func (h *Handler) PatchAnexoEstudo(w http.ResponseWriter, r *http.Request) {
	// 1. Extrair ID e validar permissões
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID da negociação inválido", http.StatusBadRequest)
		return
	}

	// 2. Decodificar o corpo da requisição (payload)
	var req PatchAnexoEstudoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if req.AnexoEstudo == "" {
		http.Error(w, "O campo 'anexoEstudo' é obrigatório", http.StatusBadRequest)
		return
	}

	// 3. Buscar negociação para garantir que existe
	var neg models.Negociacao
	if err := h.DB.First(&neg, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Negociação não encontrada", http.StatusNotFound)
			return
		}
		http.Error(w, "Erro ao buscar negociação", http.StatusInternalServerError)
		return
	}

	// Permissão: admin ou dono
	isAdmin, _ := r.Context().Value(auth.CtxIsAdmin).(bool)
	userID, _ := r.Context().Value(auth.CtxUserID).(uint)
	if !isAdmin && neg.ConsultorID != userID {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	updates := map[string]interface{}{
		"anexo_estudo": req.AnexoEstudo,
		"status":       "Estudo Feito",
	}

	if err := h.DB.Model(&neg).Updates(updates).Error; err != nil {
		http.Error(w, "Erro ao atualizar a negociação", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(neg)
}

// PatchContratoKC trata PATCH /negociacoes/{id}/contrato-kc
func (h *Handler) PatchContratoKC(w http.ResponseWriter, r *http.Request) {
	// 1. Extrair ID
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID da negociação inválido", http.StatusBadRequest)
		return
	}

	// 2. Decodificar o corpo da requisição
	var req PatchContratoKCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if req.ContratoKC == "" {
		http.Error(w, "O campo 'contratoKC' é obrigatório", http.StatusBadRequest)
		return
	}

	// 3. Buscar negociação
	var neg models.Negociacao
	if err := h.DB.First(&neg, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Negociação não encontrada", http.StatusNotFound)
			return
		}
		http.Error(w, "Erro ao buscar negociação", http.StatusInternalServerError)
		return
	}

	// Permissão: admin ou dono
	isAdmin, _ := r.Context().Value(auth.CtxIsAdmin).(bool)
	userID, _ := r.Context().Value(auth.CtxUserID).(uint)
	if !isAdmin && neg.ConsultorID != userID {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	updates := map[string]interface{}{
		"contrato_kc": req.ContratoKC,
		"status":      "Contrato Assinado",
	}

	if err := h.DB.Model(&neg).Updates(updates).Error; err != nil {
		http.Error(w, "Erro ao atualizar a negociação", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(neg)
}

// AtualizarStatus trata PATCH /negociacoes/{id}/status
func (h *Handler) AtualizarStatus(w http.ResponseWriter, r *http.Request) {
	// 1) Pega o ID da URL
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID da negociação inválido", http.StatusBadRequest)
		return
	}

	// 2) Autenticação/Autorização (admin ou dono)
	userVal := r.Context().Value(auth.CtxUserID)
	if userVal == nil {
		http.Error(w, "não autenticado", http.StatusUnauthorized)
		return
	}
	consultorID := userVal.(uint)
	isAdmin, _ := r.Context().Value(auth.CtxIsAdmin).(bool)

	// 3) Body
	type atualizarStatusRequest struct {
		Status string `json:"status"`
	}
	var payload atualizarStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if payload.Status == "" {
		http.Error(w, "o campo 'status' é obrigatório", http.StatusBadRequest)
		return
	}

	// 4) Busca a negociação pra checar dono/permissão
	neg, err := h.Repository.BuscarPorID(h.DB, uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "negociação não encontrada", http.StatusNotFound)
			return
		}
		http.Error(w, "erro ao buscar negociação", http.StatusInternalServerError)
		return
	}
	if !isAdmin && neg.ConsultorID != consultorID {
		http.Error(w, "acesso negado", http.StatusForbidden)
		return
	}

	// (Opcional) validar status permitido — comente se não quiser travar
	// allowed := map[string]bool{"Aberta":true,"Em Estudo":true,"Estudo Feito":true,"Contrato Enviado":true,"Contrato Assinado":true,"Fechada":true,"Cancelada":true}
	// if !allowed[payload.Status] { http.Error(w, "status inválido", http.StatusBadRequest); return }

	// 5) Atualiza no repositório
	if err := h.Repository.AtualizarStatus(h.DB, uint(id), payload.Status); err != nil {
		http.Error(w, "erro ao atualizar status", http.StatusInternalServerError)
		return
	}

	// 6) Retorna a negociação atualizada
	neg.Status = payload.Status
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(neg)
}
