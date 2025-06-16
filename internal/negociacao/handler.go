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
	Nome                string      `json:"nome"`
	Contato             string      `json:"contato"`
	Telefone            string      `json:"telefone"`
	CNPJ                string      `json:"cnpj"`
	Logo                string      `json:"logo"`
	AnexoFatura         string      `json:"anexoFatura"`
	AnexoEstudo         string      `json:"anexoEstudo"`
	ContratoKC          string      `json:"contratoKC"`
	AnexoContratoSocial string      `json:"anexoContratoSocial"`
	Status              string      `json:"status"`
	Produtos            string      `json:"produtos"`
	KromaTake           bool        `json:"kromaTake"`
	UF                  string      `json:"uf"`
	Arquivos            StringSlice `gorm:"type:text" json:"arquivos,omitempty"`
}

// Criar trata POST /negociacoes
func (h *Handler) Criar(w http.ResponseWriter, r *http.Request) {
	// 1) Captura consultorID do JWT
	userVal := r.Context().Value(auth.UsuarioIDKey)
	if userVal == nil {
		http.Error(w, "não autenticado", http.StatusUnauthorized)
		return
	}
	consultorID := userVal.(uint)

	// 2) Decodifica DTO
	var dto negociacaoDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// 3) Monta o modelo
	n := Negociacao{
		Nome:                dto.Nome,
		Contato:             dto.Contato,
		Telefone:            dto.Telefone,
		CNPJ:                dto.CNPJ,
		Logo:                dto.Logo,
		AnexoFatura:         dto.AnexoFatura,
		AnexoEstudo:         dto.AnexoEstudo,
		ContratoKC:          dto.ContratoKC,
		AnexoContratoSocial: dto.AnexoContratoSocial,
		Status:              dto.Status,
		Produtos:            dto.Produtos,
		KromaTake:           dto.KromaTake,
		UF:                  dto.UF,
		ConsultorID:         consultorID,
		Arquivos:            dto.Arquivos,
	}

	// 4) Persiste
	if err := h.Repository.Salvar(h.DB, &n); err != nil {
		http.Error(w, "Erro ao salvar negociação", http.StatusInternalServerError)
		return
	}

	// 5) Retorna JSON
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
	// 1) ID da URL
	idParam := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	// 2) Verifica autenticação e pega consultorID
	userVal := r.Context().Value(auth.UsuarioIDKey)
	if userVal == nil {
		http.Error(w, "não autenticado", http.StatusUnauthorized)
		return
	}
	consultorID := userVal.(uint)

	// 3) Busca registro existente
	var existing Negociacao
	if err := h.DB.First(&existing, id).Error; err != nil {
		http.Error(w, "Negociação não encontrada", http.StatusNotFound)
		return
	}

	// 4) Decodifica DTO
	var dto negociacaoDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// 5) Atualiza campos mutáveis
	existing.Nome = dto.Nome
	existing.Contato = dto.Contato
	existing.Telefone = dto.Telefone
	existing.CNPJ = dto.CNPJ
	existing.Logo = dto.Logo
	existing.AnexoFatura = dto.AnexoFatura
	existing.AnexoEstudo = dto.AnexoEstudo
	existing.ContratoKC = dto.ContratoKC
	existing.AnexoContratoSocial = dto.AnexoContratoSocial
	existing.Status = dto.Status
	existing.Produtos = dto.Produtos
	existing.KromaTake = dto.KromaTake
	existing.UF = dto.UF
	existing.ConsultorID = consultorID
	existing.Arquivos = dto.Arquivos

	// 6) Persiste atualização
	if err := h.Repository.Atualizar(h.DB, &existing); err != nil {
		http.Error(w, "Erro ao atualizar negociação", http.StatusInternalServerError)
		return
	}

	// 7) Retorna JSON
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

// Esta função adiciona novas URLs de arquivos ao slice 'Arquivos' de uma negociação existente.
func (h *Handler) AdicionarArquivos(w http.ResponseWriter, r *http.Request) {
	// 1. Pega o ID da negociação da URL.
	idParam := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "ID da negociação inválido", http.StatusBadRequest)
		return
	}

	// 2. Garante que o usuário está autenticado e pega seu ID.
	userVal := r.Context().Value(auth.UsuarioIDKey)
	if userVal == nil {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}
	consultorID := userVal.(uint)

	// 3. Decodifica o payload com as novas URLs dos arquivos.
	var req AdicionarArquivosRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	if len(req.NovosArquivos) == 0 {
		http.Error(w, "A lista 'novosArquivos' não pode estar vazia", http.StatusBadRequest)
		return
	}

	// 4. Busca a negociação existente no banco de dados.
	var negociacaoExistente Negociacao
	if err := h.DB.First(&negociacaoExistente, id).Error; err != nil {
		http.Error(w, "Negociação não encontrada", http.StatusNotFound)
		return
	}

	// 5. ✅ VERIFICAÇÃO DE SEGURANÇA:
	// Garante que o consultor logado é o "dono" desta negociação.
	isAdmin := r.Context().Value(auth.IsAdminKey).(bool)
	if !isAdmin && negociacaoExistente.ConsultorID != consultorID {
		http.Error(w, "Acesso negado: você não tem permissão para modificar esta negociação", http.StatusForbidden)
		return
	}

	// 6. Adiciona os novos arquivos ao slice existente.
	// A função 'append' do Go lida com isso de forma eficiente.
	negociacaoExistente.Arquivos = append(negociacaoExistente.Arquivos, req.NovosArquivos...)

	// 7. Salva a negociação atualizada no banco.
	if err := h.Repository.Atualizar(h.DB, &negociacaoExistente); err != nil {
		http.Error(w, "Erro ao salvar os novos arquivos", http.StatusInternalServerError)
		return
	}

	// 8. Retorna a negociação completa e atualizada.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(negociacaoExistente)
}
