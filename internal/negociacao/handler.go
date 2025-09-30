// internal/negociacao/handler.go
package negociacao

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/KromaEnergia/api-consultor/internal/auth"
	"github.com/KromaEnergia/api-consultor/internal/models"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type patchAnexoSimplesRequest struct {
	URL    string  `json:"url"`
	Status *string `json:"status"` // "Pendente" | "Enviado" | "Validado" (opcional)
}
type patchFaturaItemRequest struct {
	URL string `json:"url"`
}
type patchFaturaStatusRequest struct {
	Status string `json:"status"`
}

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

// se status não vier ou vier vazio, usamos "Enviado"
func statusOrDefaultEnviado(s *string) string {
	if s == nil || strings.TrimSpace(*s) == "" {
		return models.StatusEnviado
	}
	return strings.TrimSpace(*s)
}

// NewHandler cria um novo handler de negociações
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		DB:         db,
		Repository: NewRepository(),
	}
}

// ================== DTOs LOCAIS (no mesmo arquivo) ==================

type multiAnexoDTO struct {
	Itens  []string `json:"itens"`            // links da fatura
	Status string   `json:"status,omitempty"` // ex.: "Pendente", "Enviado", "Validado"
}

type negociacaoCreateDTO struct {
	// Básico
	Nome            string `json:"nome"`
	Email           string `json:"email"`
	Contato         string `json:"contato"`
	NumeroDoContato string `json:"numeroDoContato"`
	Telefone        string `json:"telefone"`
	CNPJ            string `json:"cnpj"`
	UF              string `json:"uf"`
	Status          string `json:"status"`
	KromaTake       bool   `json:"kromaTake"`

	// anexos simples + status (STRING)
	Logo                      string `json:"logo"`
	LogoStatus                string `json:"logoStatus"`
	AnexoEstudo               string `json:"anexoEstudo"`
	AnexoEstudoStatus         string `json:"anexoEstudoStatus"`
	ContratoKC                string `json:"contratoKC"`
	ContratoKCStatus          string `json:"contratoKCStatus"`
	AnexoContratoSocial       string `json:"anexoContratoSocial"`
	AnexoContratoSocialStatus string `json:"anexoContratoSocialStatus"`

	// novos anexos + status (STRING)
	AnexoProcuracao                string `json:"anexoProcuracao"`
	AnexoProcuracaoStatus          string `json:"anexoProcuracaoStatus"`
	AnexoRepresentanteLegal        string `json:"anexoRepresentanteLegal"`
	AnexoRepresentanteLegalStatus  string `json:"anexoRepresentanteLegalStatus"`
	AnexoEstudoDeViabilidade       string `json:"anexoEstudoDeViabilidade"`
	AnexoEstudoDeViabilidadeStatus string `json:"anexoEstudoDeViabilidadeStatus"`

	// Fatura
	AnexoFatura multiAnexoDTO `json:"anexoFatura"`

	// Outros anexos soltos
	Arquivos []string `json:"arquivos"`
}

/* ================== POST /negociacoes (Criar) ================== */
type negociacaoUpdateDTO struct {
	// básicos
	Nome            string `json:"nome"`
	Email           string `json:"email"`
	Contato         string `json:"contato"`
	NumeroDoContato string `json:"numeroDoContato"`
	Telefone        string `json:"telefone"`
	CNPJ            string `json:"cnpj"`
	UF              string `json:"uf"`
	Status          string `json:"status"`
	KromaTake       bool   `json:"kromaTake"`

	// anexos simples + status (STRING)
	Logo                      string `json:"logo"`
	LogoStatus                string `json:"logoStatus"`
	AnexoEstudo               string `json:"anexoEstudo"`
	AnexoEstudoStatus         string `json:"anexoEstudoStatus"`
	ContratoKC                string `json:"contratoKC"`
	ContratoKCStatus          string `json:"contratoKCStatus"`
	AnexoContratoSocial       string `json:"anexoContratoSocial"`
	AnexoContratoSocialStatus string `json:"anexoContratoSocialStatus"`

	// novos anexos + status (STRING)
	AnexoProcuracao                string `json:"anexoProcuracao"`
	AnexoProcuracaoStatus          string `json:"anexoProcuracaoStatus"`
	AnexoRepresentanteLegal        string `json:"anexoRepresentanteLegal"`
	AnexoRepresentanteLegalStatus  string `json:"anexoRepresentanteLegalStatus"`
	AnexoEstudoDeViabilidade       string `json:"anexoEstudoDeViabilidade"`
	AnexoEstudoDeViabilidadeStatus string `json:"anexoEstudoDeViabilidadeStatus"`

	// fatura
	AnexoFatura multiAnexoDTO `json:"anexoFatura"`

	// anexos soltos
	Arquivos []string `json:"arquivos"`
}

// Aceita tanto "anexoFatura": "https://..." quanto
// "anexoFatura": { "itens": ["https://..."], "status": "Pendente" }
func (m *multiAnexoDTO) UnmarshalJSON(b []byte) error {
	// Caso seja uma string: "https://..."
	if len(b) > 0 && b[0] == '"' {
		var url string
		if err := json.Unmarshal(b, &url); err != nil {
			return err
		}
		m.Itens = []string{}
		if strings.TrimSpace(url) != "" {
			m.Itens = []string{url}
		}
		// status fica vazio; você pode setar default depois ("Pendente")
		return nil
	}

	// Caso seja objeto: {"itens":[...], "status":"..."}
	type alias multiAnexoDTO
	var aux alias
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	*m = multiAnexoDTO(aux)
	return nil
}

func (h *Handler) Criar(w http.ResponseWriter, r *http.Request) {
	userVal := r.Context().Value(auth.CtxUserID)
	if userVal == nil {
		http.Error(w, "não autenticado", http.StatusUnauthorized)
		return
	}
	consultorID := userVal.(uint)

	var dto negociacaoCreateDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if dto.Arquivos == nil {
		dto.Arquivos = []string{}
	}

	// Default de status, se não vier
	if strings.TrimSpace(dto.Status) == "" {
		dto.Status = "Pendente"
	}

	// Normaliza o anexoFatura vindo do DTO (já pode ter vindo como string -> itens[0])
	anexoFatura := models.MultiAnexo{
		Itens:  dto.AnexoFatura.Itens,
		Status: dto.AnexoFatura.Status,
	}
	if len(anexoFatura.Itens) == 0 && anexoFatura.Status == "" {
		// nenhum anexo enviado; ok manter vazio
	}
	if anexoFatura.Status == "" && len(anexoFatura.Itens) > 0 {
		// se veio só a URL em formato string, seta status default
		anexoFatura.Status = "Pendente"
	}

	n := models.Negociacao{
		Nome:            dto.Nome,
		Email:           dto.Email,
		Contato:         dto.Contato,
		NumeroDoContato: dto.NumeroDoContato,
		Telefone:        dto.Telefone,
		CNPJ:            dto.CNPJ,
		UF:              dto.UF,
		Status:          dto.Status,
		KromaTake:       dto.KromaTake,

		// Fatura (objeto)
		AnexoFatura: anexoFatura,

		// Outros anexos soltos
		Arquivos:    dto.Arquivos,
		ConsultorID: consultorID,
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

// ================== PUT /negociacoes/{id} ==================

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

	// Decodifica JSON com o shape correto
	var dto negociacaoUpdateDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Helpers de normalização
	def := func(s string) string {
		if strings.TrimSpace(s) == "" {
			return models.StatusPendente
		}
		return strings.TrimSpace(s)
	}

	// Defaults de slices
	if dto.Arquivos == nil {
		dto.Arquivos = []string{}
	}

	// Normaliza o objeto de fatura
	anexoFatura := models.MultiAnexo{
		Itens:  dto.AnexoFatura.Itens,
		Status: strings.TrimSpace(dto.AnexoFatura.Status),
	}
	// Se vier itens e status em branco, deixa "Pendente" (o upload setará "Enviado")
	if len(anexoFatura.Itens) > 0 && anexoFatura.Status == "" {
		anexoFatura.Status = models.StatusPendente
	}

	// Mapeia campos básicos
	existing.Nome = dto.Nome
	existing.Contato = dto.Contato
	existing.Email = dto.Email
	existing.NumeroDoContato = dto.NumeroDoContato
	existing.Telefone = dto.Telefone
	existing.CNPJ = dto.CNPJ
	existing.UF = dto.UF
	existing.Status = dto.Status
	existing.KromaTake = dto.KromaTake

	// Anexos simples + status (string) com defaults "Pendente" se vierem vazios
	existing.Logo = dto.Logo
	existing.LogoStatus = def(dto.LogoStatus)

	existing.AnexoEstudo = dto.AnexoEstudo
	existing.AnexoEstudoStatus = def(dto.AnexoEstudoStatus)

	existing.ContratoKC = dto.ContratoKC
	existing.ContratoKCStatus = def(dto.ContratoKCStatus)

	existing.AnexoContratoSocial = dto.AnexoContratoSocial
	existing.AnexoContratoSocialStatus = def(dto.AnexoContratoSocialStatus)

	// Novos anexos + status
	existing.AnexoProcuracao = dto.AnexoProcuracao
	existing.AnexoProcuracaoStatus = def(dto.AnexoProcuracaoStatus)

	existing.AnexoRepresentanteLegal = dto.AnexoRepresentanteLegal
	existing.AnexoRepresentanteLegalStatus = def(dto.AnexoRepresentanteLegalStatus)

	existing.AnexoEstudoDeViabilidade = dto.AnexoEstudoDeViabilidade
	existing.AnexoEstudoDeViabilidadeStatus = def(dto.AnexoEstudoDeViabilidadeStatus)

	// Fatura (objeto jsonb)
	existing.AnexoFatura = anexoFatura

	// Outros anexos soltos
	existing.Arquivos = dto.Arquivos

	// mantém o dono como quem está atualizando
	existing.ConsultorID = consultorID

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

/* ================== PATCH /negociacoes/{id}/logo ================== */
func (h *Handler) PatchLogo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var req patchAnexoSimplesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	var neg models.Negociacao
	if err := h.DB.First(&neg, id).Error; err != nil {
		http.Error(w, "Negociação não encontrada", http.StatusNotFound)
		return
	}
	isAdmin, _ := r.Context().Value(auth.CtxIsAdmin).(bool)
	userID, _ := r.Context().Value(auth.CtxUserID).(uint)
	if !isAdmin && neg.ConsultorID != userID {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	status := statusOrDefaultEnviado(req.Status)
	updates := map[string]any{
		"logo":        req.URL,
		"logo_status": status,
	}
	if err := h.DB.Model(&neg).Updates(updates).Error; err != nil {
		http.Error(w, "Erro ao atualizar logo", http.StatusInternalServerError)
		return
	}
	neg.Logo = req.URL
	neg.LogoStatus = status

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(neg)
}

func (h *Handler) PatchAnexoContratoSocial(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var req patchAnexoSimplesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	var neg models.Negociacao
	if err := h.DB.First(&neg, id).Error; err != nil {
		http.Error(w, "Negociação não encontrada", http.StatusNotFound)
		return
	}
	isAdmin, _ := r.Context().Value(auth.CtxIsAdmin).(bool)
	userID, _ := r.Context().Value(auth.CtxUserID).(uint)
	if !isAdmin && neg.ConsultorID != userID {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	status := statusOrDefaultEnviado(req.Status)
	updates := map[string]any{
		"anexo_contrato_social":        req.URL,
		"anexo_contrato_social_status": status,
	}
	if err := h.DB.Model(&neg).Updates(updates).Error; err != nil {
		http.Error(w, "Erro ao atualizar contrato social", http.StatusInternalServerError)
		return
	}
	neg.AnexoContratoSocial = req.URL
	neg.AnexoContratoSocialStatus = status

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(neg)
}

func (h *Handler) PatchAnexoProcuracao(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var req patchAnexoSimplesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	var neg models.Negociacao
	if err := h.DB.First(&neg, id).Error; err != nil {
		http.Error(w, "Negociação não encontrada", http.StatusNotFound)
		return
	}
	isAdmin, _ := r.Context().Value(auth.CtxIsAdmin).(bool)
	userID, _ := r.Context().Value(auth.CtxUserID).(uint)
	if !isAdmin && neg.ConsultorID != userID {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	status := statusOrDefaultEnviado(req.Status)
	updates := map[string]any{
		"anexo_procuracao":        req.URL,
		"anexo_procuracao_status": status,
	}
	if err := h.DB.Model(&neg).Updates(updates).Error; err != nil {
		http.Error(w, "Erro ao atualizar procuração", http.StatusInternalServerError)
		return
	}
	neg.AnexoProcuracao = req.URL
	neg.AnexoProcuracaoStatus = status

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(neg)
}

func (h *Handler) PatchAnexoRepresentanteLegal(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var req patchAnexoSimplesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	var neg models.Negociacao
	if err := h.DB.First(&neg, id).Error; err != nil {
		http.Error(w, "Negociação não encontrada", http.StatusNotFound)
		return
	}
	isAdmin, _ := r.Context().Value(auth.CtxIsAdmin).(bool)
	userID, _ := r.Context().Value(auth.CtxUserID).(uint)
	if !isAdmin && neg.ConsultorID != userID {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	status := statusOrDefaultEnviado(req.Status)
	updates := map[string]any{
		"anexo_representante_legal":        req.URL,
		"anexo_representante_legal_status": status,
	}
	if err := h.DB.Model(&neg).Updates(updates).Error; err != nil {
		http.Error(w, "Erro ao atualizar representante legal", http.StatusInternalServerError)
		return
	}
	neg.AnexoRepresentanteLegal = req.URL
	neg.AnexoRepresentanteLegalStatus = status

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(neg)
}

func (h *Handler) PatchAnexoEstudoDeViabilidade(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var req patchAnexoSimplesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	var neg models.Negociacao
	if err := h.DB.First(&neg, id).Error; err != nil {
		http.Error(w, "Negociação não encontrada", http.StatusNotFound)
		return
	}
	isAdmin, _ := r.Context().Value(auth.CtxIsAdmin).(bool)
	userID, _ := r.Context().Value(auth.CtxUserID).(uint)
	if !isAdmin && neg.ConsultorID != userID {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	status := statusOrDefaultEnviado(req.Status)
	updates := map[string]any{
		"anexo_estudo_de_viabilidade":        req.URL,
		"anexo_estudo_de_viabilidade_status": status,
	}
	if err := h.DB.Model(&neg).Updates(updates).Error; err != nil {
		http.Error(w, "Erro ao atualizar estudo de viabilidade", http.StatusInternalServerError)
		return
	}
	neg.AnexoEstudoDeViabilidade = req.URL
	neg.AnexoEstudoDeViabilidadeStatus = status

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(neg)
}

func (h *Handler) PostFaturaItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var req patchFaturaItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.URL) == "" {
		http.Error(w, "JSON inválido ou URL vazia", http.StatusBadRequest)
		return
	}

	var neg models.Negociacao
	if err := h.DB.First(&neg, id).Error; err != nil {
		http.Error(w, "Negociação não encontrada", http.StatusNotFound)
		return
	}
	isAdmin, _ := r.Context().Value(auth.CtxIsAdmin).(bool)
	userID, _ := r.Context().Value(auth.CtxUserID).(uint)
	if !isAdmin && neg.ConsultorID != userID {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	// evita duplicado (opcional)
	exists := false
	for _, it := range neg.AnexoFatura.Itens {
		if it == req.URL {
			exists = true
			break
		}
	}
	if !exists {
		neg.AnexoFatura.Itens = append(neg.AnexoFatura.Itens, req.URL)
	}

	// se sem status ou "Pendente", marca "Enviado"
	if strings.TrimSpace(neg.AnexoFatura.Status) == "" ||
		strings.EqualFold(neg.AnexoFatura.Status, models.StatusPendente) {
		neg.AnexoFatura.Status = models.StatusEnviado
	}

	if err := h.DB.Model(&neg).Update("anexo_fatura", neg.AnexoFatura).Error; err != nil {
		http.Error(w, "Erro ao adicionar item de fatura", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(neg.AnexoFatura)
}

func (h *Handler) DeleteFaturaItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	idx, err := strconv.Atoi(vars["idx"])
	if err != nil || idx < 0 {
		http.Error(w, "Índice inválido", http.StatusBadRequest)
		return
	}

	var neg models.Negociacao
	if err := h.DB.First(&neg, id).Error; err != nil {
		http.Error(w, "Negociação não encontrada", http.StatusNotFound)
		return
	}
	isAdmin, _ := r.Context().Value(auth.CtxIsAdmin).(bool)
	userID, _ := r.Context().Value(auth.CtxUserID).(uint)
	if !isAdmin && neg.ConsultorID != userID {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	if idx >= len(neg.AnexoFatura.Itens) {
		http.Error(w, "Índice fora do intervalo", http.StatusBadRequest)
		return
	}
	neg.AnexoFatura.Itens = append(neg.AnexoFatura.Itens[:idx], neg.AnexoFatura.Itens[idx+1:]...)
	if err := h.DB.Model(&neg).Update("anexo_fatura", neg.AnexoFatura).Error; err != nil {
		http.Error(w, "Erro ao remover item de fatura", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(neg.AnexoFatura)
}

func (h *Handler) PatchFaturaStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var req patchFaturaStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	var neg models.Negociacao
	if err := h.DB.First(&neg, id).Error; err != nil {
		http.Error(w, "Negociação não encontrada", http.StatusNotFound)
		return
	}
	isAdmin, _ := r.Context().Value(auth.CtxIsAdmin).(bool)
	userID, _ := r.Context().Value(auth.CtxUserID).(uint)
	if !isAdmin && neg.ConsultorID != userID {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	neg.AnexoFatura.Status = strings.TrimSpace(req.Status)
	if err := h.DB.Model(&neg).Update("anexo_fatura", neg.AnexoFatura).Error; err != nil {
		http.Error(w, "Erro ao atualizar status da fatura", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(neg.AnexoFatura)
}
