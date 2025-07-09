// internal/produtos/handler.go
package produtos

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Handler struct {
	Repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{Repo: repo}
}

// POST /negociacoes/{id}/produtos
func (h *Handler) CreateProdutos(w http.ResponseWriter, r *http.Request) {
	negID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID de negociação inválido", http.StatusBadRequest)
		return
	}

	var body struct {
		Produtos []Produto `json:"produtos"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "JSON mal formado", http.StatusBadRequest)
		return
	}

	for i := range body.Produtos {
		body.Produtos[i].NegociacaoID = uint(negID)
	}

	if err := h.Repo.CreateMany(body.Produtos); err != nil {
		http.Error(w, "Erro ao inserir produtos", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(body.Produtos)
}

// GET /negociacoes/{id}/produtos
func (h *Handler) ListProdutos(w http.ResponseWriter, r *http.Request) {
	negID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID de negociação inválido", http.StatusBadRequest)
		return
	}

	produtos, err := h.Repo.FindByNeg(uint(negID))
	if err != nil {
		http.Error(w, "Erro ao buscar produtos", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(produtos)
}

// GET /negociacoes/{id}/produtos/{pid}
func (h *Handler) GetProduto(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(mux.Vars(r)["pid"])
	if err != nil {
		http.Error(w, "ID de produto inválido", http.StatusBadRequest)
		return
	}

	prod, err := h.Repo.FindByID(uint(pid))
	if err != nil {
		http.Error(w, "Produto não encontrado", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(prod)
}

// PUT /negociacoes/{id}/produtos/{pid}
func (h *Handler) UpdateProduto(w http.ResponseWriter, r *http.Request) {
	negID, err1 := strconv.Atoi(mux.Vars(r)["id"])
	pid, err2 := strconv.Atoi(mux.Vars(r)["pid"])
	if err1 != nil || err2 != nil {
		http.Error(w, "IDs inválidos", http.StatusBadRequest)
		return
	}

	existing, err := h.Repo.FindByID(uint(pid))
	if err != nil || existing.NegociacaoID != uint(negID) {
		http.Error(w, "Produto não encontrado para essa negociação", http.StatusNotFound)
		return
	}

	var body Produto
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "JSON mal formado", http.StatusBadRequest)
		return
	}

	// atualiza campos
	existing.Tipo = body.Tipo
	existing.Ativo = body.Ativo
	existing.Comissao = body.Comissao
	existing.Fee = body.Fee

	if err := h.Repo.Update(existing); err != nil {
		http.Error(w, "Erro ao atualizar produto", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(existing)
}

// DELETE /negociacoes/{id}/produtos/{pid}
func (h *Handler) DeleteProduto(w http.ResponseWriter, r *http.Request) {
	negID, err1 := strconv.Atoi(mux.Vars(r)["id"])
	pid, err2 := strconv.Atoi(mux.Vars(r)["pid"])
	if err1 != nil || err2 != nil {
		http.Error(w, "IDs inválidos", http.StatusBadRequest)
		return
	}

	existing, err := h.Repo.FindByID(uint(pid))
	if err != nil || existing.NegociacaoID != uint(negID) {
		http.Error(w, "Produto não encontrado para essa negociação", http.StatusNotFound)
		return
	}

	if err := h.Repo.Delete(existing); err != nil {
		http.Error(w, "Erro ao deletar produto", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
