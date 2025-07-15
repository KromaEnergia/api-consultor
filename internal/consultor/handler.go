package consultor

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/KromaEnergia/api-consultor/internal/auth"
	"github.com/KromaEnergia/api-consultor/internal/contrato"
	"github.com/KromaEnergia/api-consultor/internal/negociacao"
	"github.com/KromaEnergia/api-consultor/internal/produtos"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// DTOs

// ResumoComissoes agrupa valores de energia e gestão por status
type ResumoComissoes struct {
	ComissoesRecebidas struct {
		Energia float64 `json:"energiaRecebida"`
		Gestao  float64 `json:"gestaoRecebida"`
	} `json:"comissoesRecebidas"`
	ComissoesAReceber struct {
		Energia float64 `json:"energiaAReceber"`
		Gestao  float64 `json:"gestaoAReceber"`
	} `json:"comissoesAReceber"`
}

type SolicitacaoCNPJRequest struct {
	NovoCNPJ string `json:"novoCnpj"`
}

type AprovarCNPJRequest struct {
	Aprovado bool `json:"aprovado"`
}

type AtualizarTermoRequest struct {
	URL string `json:"url"`
}
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
type SolicitacaoEmailRequest struct {
	NovoEmail string `json:"novoEmail"`
}

type AprovarEmailRequest struct {
	Aprovado bool `json:"aprovado"`
}

type createConsultorRequest struct {
	Nome            string     `json:"nome"`
	Sobrenome       string     `json:"sobrenome"`
	CNPJ            string     `json:"cnpj"`
	Email           string     `json:"email"`
	Telefone        string     `json:"telefone"`
	Foto            string     `json:"foto"`
	TermoDeParceria string     `json:"termoDeParceria"`
	DataNascimento  CustomDate `json:"dataNascimento"`
	Estado          string     `json:"estado"`
	Senha           string     `json:"senha"`
	IsAdmin         bool       `json:"isAdmin"`
	ComercialID     uint       `json:"comercial_id"` // ← aqui
}

// Handler encapsula DB e repo
type Handler struct {
	DB         *gorm.DB
	Repository Repository
}

// ComissoesHandler trata a rota de resumo de comissões
type ComissoesHandler struct {
	DB *gorm.DB
}

// NewComissoesHandler cria um handler para comissões
func NewComissoesHandler(db *gorm.DB) *ComissoesHandler {
	return &ComissoesHandler{DB: db}
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

	// usa req.Login (que já contém o e-mail vindo do JSON)
	user, err := h.Repository.BuscarPorEmail(h.DB, req.Login)
	if err != nil {
		http.Error(w, "credenciais inválidas", http.StatusUnauthorized)
		return
	}
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

	// 1. Opcional: validar que ComercialID != 0
	if req.ComercialID == 0 {
		http.Error(w, "comercial_id é obrigatório", http.StatusBadRequest)
		return
	}

	// 2. Hash da senha
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Senha), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "erro ao processar senha", http.StatusInternalServerError)
		return
	}

	// 3. Preencher o struct incluindo o ComercialID
	c := Consultor{
		Nome:                  req.Nome,
		Sobrenome:             req.Sobrenome,
		CNPJ:                  req.CNPJ,
		Email:                 req.Email,
		Telefone:              req.Telefone,
		Foto:                  req.Foto,
		TermoDeParceria:       req.TermoDeParceria,
		DataNascimento:        req.DataNascimento,
		Estado:                req.Estado,
		Senha:                 string(hash),
		PrecisaRedefinirSenha: false,
		IsAdmin:               req.IsAdmin,
		ComercialID:           req.ComercialID, // ← aqui
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
	prodRepo := produtos.NewRepository(h.DB)
	produtosList, err := prodRepo.ListarPorConsultor(consultorObj.ID)
	dto := MontarResumoConsultorDTO(*consultorObj, contratos, negociacoes, produtosList)

	w.Header().Set("Content-Type", "application/json")
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
		Preload("Negociacoes.Produtos").
		Preload("Negociacoes.Comentarios").
		Preload("Negociacoes.CalculosComissao").
		Preload("Negociacoes.CalculosComissao.Parcelas").
		First(&c, userID).Error; err != nil {
		http.Error(w, "Consultor não encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

// SolicitarAlteracaoCNPJ permite que um consultor peça a mudança do seu CNPJ
func (h *Handler) SolicitarAlteracaoCNPJ(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UsuarioIDKey).(uint)
	isAdmin := r.Context().Value(auth.IsAdminKey).(bool)

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	// Um consultor só pode solicitar para si mesmo
	if !isAdmin && uint(id) != userID {
		http.Error(w, "acesso negado", http.StatusForbidden)
		return
	}

	var req SolicitacaoCNPJRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "payload inválido", http.StatusBadRequest)
		return
	}

	if req.NovoCNPJ == "" {
		http.Error(w, "o campo 'novoCnpj' é obrigatório", http.StatusBadRequest)
		return
	}

	// Busca o consultor para atualizar
	consultor, err := h.Repository.BuscarPorID(h.DB, uint(id))
	if err != nil {
		http.Error(w, "consultor não encontrado", http.StatusNotFound)
		return
	}

	// Atualiza apenas os campos da solicitação
	consultor.RequestedCNPJ = req.NovoCNPJ
	consultor.CNPJChangeApproved = false // Reseta a aprovação a cada nova solicitação

	if err := h.Repository.Salvar(h.DB, consultor); err != nil {
		http.Error(w, "erro ao salvar solicitação", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Solicitação de alteração de CNPJ enviada para aprovação."})
}

// GerenciarAlteracaoCNPJ permite que um admin aprove ou negue a mudança de CNPJ
func (h *Handler) GerenciarAlteracaoCNPJ(w http.ResponseWriter, r *http.Request) {
	isAdmin := r.Context().Value(auth.IsAdminKey).(bool)

	// Apenas admins podem aprovar/negar
	if !isAdmin {
		http.Error(w, "acesso negado, rota exclusiva para administradores", http.StatusForbidden)
		return
	}

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var req AprovarCNPJRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "payload inválido", http.StatusBadRequest)
		return
	}

	consultor, err := h.Repository.BuscarPorID(h.DB, uint(id))
	if err != nil {
		http.Error(w, "consultor não encontrado", http.StatusNotFound)
		return
	}

	if consultor.RequestedCNPJ == "" {
		http.Error(w, "não há solicitação de CNPJ pendente para este consultor", http.StatusBadRequest)
		return
	}

	if req.Aprovado {
		// Se aprovado, atualiza o CNPJ principal e limpa a solicitação
		consultor.CNPJ = consultor.RequestedCNPJ
		consultor.RequestedCNPJ = ""
		consultor.CNPJChangeApproved = true
	} else {
		// Se negado, apenas limpa a solicitação
		consultor.RequestedCNPJ = ""
		consultor.CNPJChangeApproved = false
	}

	if err := h.Repository.Salvar(h.DB, consultor); err != nil {
		http.Error(w, "erro ao processar a solicitação", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Solicitação de alteração de CNPJ gerenciada com sucesso."})
}

// AtualizarTermoDeParceria permite que um consultor adicione/atualize seu link do termo
func (h *Handler) AtualizarTermoDeParceria(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UsuarioIDKey).(uint)
	isAdmin := r.Context().Value(auth.IsAdminKey).(bool)

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	// Um consultor só pode atualizar o seu próprio termo
	if !isAdmin && uint(id) != userID {
		http.Error(w, "acesso negado", http.StatusForbidden)
		return
	}

	var req AtualizarTermoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "payload inválido", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "o campo 'url' é obrigatório", http.StatusBadRequest)
		return
	}

	consultor, err := h.Repository.BuscarPorID(h.DB, uint(id))
	if err != nil {
		http.Error(w, "consultor não encontrado", http.StatusNotFound)
		return
	}

	// Atualiza apenas o termo de parceria
	consultor.TermoDeParceria = req.URL

	if err := h.Repository.Salvar(h.DB, consultor); err != nil {
		http.Error(w, "erro ao salvar termo de parceria", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Termo de parceria atualizado com sucesso."})
}

// Em seu handler.go

func (h *Handler) AtualizarMeuPerfil(w http.ResponseWriter, r *http.Request) {
	// 1. Pega o ID do usuário que vem do token de autenticação
	userID, ok := r.Context().Value(auth.UsuarioIDKey).(uint)
	if !ok {
		http.Error(w, "ID de usuário inválido no token", http.StatusUnauthorized)
		return
	}

	// 2. Busca o registro ATUAL do consultor no banco de dados
	var consultorExistente Consultor
	if err := h.DB.First(&consultorExistente, userID).Error; err != nil {
		http.Error(w, "Consultor não encontrado", http.StatusNotFound)
		return
	}

	// 3. Decodifica os novos dados POR CIMA do registro existente.
	// Campos não enviados no JSON (como CNPJ) não serão alterados.
	if err := json.NewDecoder(r.Body).Decode(&consultorExistente); err != nil {
		http.Error(w, "Payload inválido", http.StatusBadRequest)
		return
	}

	// 4. Salva o objeto completo e atualizado de volta no banco
	if err := h.DB.Save(&consultorExistente).Error; err != nil {
		http.Error(w, "Erro ao atualizar o perfil", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(consultorExistente) // Retorna o perfil atualizado
}

// SolicitarAlteracaoEmail permite que um consultor peça a mudança do seu e-mail.
func (h *Handler) SolicitarAlteracaoEmail(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UsuarioIDKey).(uint)
	isAdmin := r.Context().Value(auth.IsAdminKey).(bool)

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	// Um consultor só pode solicitar para si mesmo (a menos que seja admin).
	if !isAdmin && uint(id) != userID {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	var req SolicitacaoEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Payload inválido", http.StatusBadRequest)
		return
	}

	if req.NovoEmail == "" {
		http.Error(w, "O campo 'novoEmail' é obrigatório", http.StatusBadRequest)
		return
	}

	// Busca o consultor para atualizar.
	consultor, err := h.Repository.BuscarPorID(h.DB, uint(id))
	if err != nil {
		http.Error(w, "Consultor não encontrado", http.StatusNotFound)
		return
	}

	// Atualiza os campos da solicitação de e-mail.
	consultor.RequestedEmail = req.NovoEmail
	consultor.EmailChangeApproved = false // Reseta a aprovação a cada nova solicitação.

	if err := h.Repository.Salvar(h.DB, consultor); err != nil {
		http.Error(w, "Erro ao salvar solicitação de e-mail", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Solicitação de alteração de e-mail enviada para aprovação."})
}

// GerenciarAlteracaoEmail permite que um admin aprove ou negue a mudança de e-mail.
func (h *Handler) GerenciarAlteracaoEmail(w http.ResponseWriter, r *http.Request) {
	isAdmin := r.Context().Value(auth.IsAdminKey).(bool)

	// Apenas admins podem aprovar/negar.
	if !isAdmin {
		http.Error(w, "Acesso negado, rota exclusiva para administradores", http.StatusForbidden)
		return
	}

	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var req AprovarEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Payload inválido", http.StatusBadRequest)
		return
	}

	consultor, err := h.Repository.BuscarPorID(h.DB, uint(id))
	if err != nil {
		http.Error(w, "Consultor não encontrado", http.StatusNotFound)
		return
	}

	if consultor.RequestedEmail == "" {
		http.Error(w, "Não há solicitação de e-mail pendente para este consultor", http.StatusBadRequest)
		return
	}

	if req.Aprovado {
		// Se aprovado, atualiza o e-mail principal e limpa a solicitação.
		consultor.Email = consultor.RequestedEmail
		consultor.RequestedEmail = ""
		consultor.EmailChangeApproved = true
	} else {
		// Se negado, apenas limpa a solicitação.
		consultor.RequestedEmail = ""
		consultor.EmailChangeApproved = false
	}

	if err := h.Repository.Salvar(h.DB, consultor); err != nil {
		http.Error(w, "Erro ao processar a solicitação de e-mail", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Solicitação de alteração de e-mail gerenciada com sucesso."})
}

// GetResumo trata GET /consultores/comissoes usando o usuário autenticado
func (h *ComissoesHandler) GetResumo(w http.ResponseWriter, r *http.Request) {
	// Extrai ID do consultor do contexto (do JWT)
	userID := r.Context().Value(auth.UsuarioIDKey).(uint)

	var res ResumoComissoes

	// Total recebido (status Pago) para negociações fechadas
	h.DB.Table("parcela_comissaos").
		Select(
			"SUM(calculo_comissaos.energia_mensal) AS energia_recebida, "+
				"SUM(calculo_comissaos.valor_gestao_mensal) AS gestao_recebida").
		Joins("JOIN calculo_comissaos ON calculo_comissaos.id = parcela_comissaos.calculo_comissao_id").
		Joins("JOIN negociacoes ON negociacoes.id = calculo_comissaos.negociacao_id").
		Where("negociacoes.consultor_id = ? AND negociacoes.status = ? AND parcela_comissaos.status = ?", userID, "Fechada", "Pago").
		Scan(&res.ComissoesRecebidas)

	// Total a receber (Pendente ou Atrasado)
	h.DB.Table("parcela_comissaos").
		Select(
			"SUM(calculo_comissaos.energia_mensal) AS energia_a_receber, "+
				"SUM(calculo_comissaos.valor_gestao_mensal) AS gestao_a_receber").
		Joins("JOIN calculo_comissaos ON calculo_comissaos.id = parcela_comissaos.calculo_comissao_id").
		Joins("JOIN negociacoes ON negociacoes.id = calculo_comissaos.negociacao_id").
		Where("negociacoes.consultor_id = ? AND negociacoes.status = ? AND parcela_comissaos.status IN ?", userID, "Fechada", []string{"Pendente", "Atrasado"}).
		Scan(&res.ComissoesAReceber)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
