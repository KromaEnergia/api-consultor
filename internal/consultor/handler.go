package consultor

import (
	"encoding/json"
	"net/http"
	"strconv"
"api/internal/contrato"
	"api/internal/negociacao"
	"github.com/gorilla/mux"
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

// POST /consultores
func (h *Handler) CriarConsultor(w http.ResponseWriter, r *http.Request) {
	var c Consultor
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}
	if err := h.Repository.Salvar(h.DB, &c); err != nil {
		http.Error(w, "Erro ao salvar consultor", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

// GET /consultores
func (h *Handler) ListarConsultores(w http.ResponseWriter, r *http.Request) {
	consultores, err := h.Repository.ListarTodos(h.DB)
	if err != nil {
		http.Error(w, "Erro ao listar consultores", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(consultores)
}

// GET /consultores/{id}
func (h *Handler) BuscarPorID(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	consultor, err := h.Repository.BuscarPorID(h.DB, uint(id))
	if err != nil {
		http.Error(w, "Consultor não encontrado", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(consultor)
}

// PUT /consultores/{id}
func (h *Handler) AtualizarConsultor(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	var dadosAtualizados Consultor
	if err := json.NewDecoder(r.Body).Decode(&dadosAtualizados); err != nil {
		http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
		return
	}
	if err := h.Repository.Atualizar(h.DB, uint(id), &dadosAtualizados); err != nil {
		http.Error(w, "Erro ao atualizar consultor", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Consultor atualizado com sucesso"))
}

// DELETE /consultores/{id}
func (h *Handler) DeletarConsultor(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	if err := h.Repository.Deletar(h.DB, uint(id)); err != nil {
		http.Error(w, "Erro ao excluir consultor", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Consultor excluído com sucesso"))
}
func (h *Handler) ObterResumoConsultor(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idStr)

	consultor, err := h.Repository.BuscarPorID(h.DB, uint(id))
	if err != nil {
		http.Error(w, "Consultor não encontrado", http.StatusNotFound)
		return
	}

	negociacoes, _ := negociacao.NewRepository().ListarPorConsultor(h.DB, consultor.ID)
	contratos, _ := contrato.NewRepository().ListarPorConsultor(h.DB, consultor.ID)

	dto := MontarResumoConsultorDTO(*consultor, contratos, negociacoes)

	json.NewEncoder(w).Encode(dto)
}
