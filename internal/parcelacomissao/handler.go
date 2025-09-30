package parcelacomissao

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

/* ============================== Handler & DTOs ============================== */

type Handler struct {
	Repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{Repo: repo}
}

// DTO usado no PUT /parcelas/{pid}
type ParcelaUpdateDTO struct {
	Valor          float64   `json:"valor"`
	DataVencimento time.Time `json:"dataVencimento"`
	Status         string    `json:"status"`
	Anexo          string    `json:"anexo"`
	NotaFiscal     string    `json:"notaFiscal"`
	VolumeMensal   float64   `json:"volumeMensal"`
}

// DTO usado no POST /calculos-comissao/{cid}/parcelas
type ParcelaCreateDTO struct {
	Valor          float64   `json:"valor"`
	DataVencimento time.Time `json:"dataVencimento"`
	Status         string    `json:"status"` // se vazio, assume "Pendente"
	Anexo          string    `json:"anexo"`  // opcional
	NotaFiscal     string    `json:"notaFiscal"`
	VolumeMensal   float64   `json:"volumeMensal"` // opcional
}

/* ============================== Utilidades ============================== */

// Soma as parcelas do cálculo e atualiza calculo_comissaos.total_receber
func recalcTotalForCalculo(db *gorm.DB, calculoID uint) error {
	var total float64
	if err := db.Model(&ParcelaComissao{}).
		Where("calculo_comissao_id = ?", calculoID).
		Select("COALESCE(SUM(valor), 0)").
		Scan(&total).Error; err != nil {
		return err
	}
	return db.Table("calculo_comissaos").
		Where("id = ?", calculoID).
		Update("total_receber", total).Error
}

/* ============================== Endpoints ============================== */

// GET /calculos-comissao/{cid}/parcelas
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	cid, err := strconv.Atoi(mux.Vars(r)["cid"])
	if err != nil {
		http.Error(w, "ID do cálculo de comissão inválido", http.StatusBadRequest)
		return
	}

	parcelas, err := h.Repo.ListByCalculoID(uint(cid))
	if err != nil {
		http.Error(w, "Erro ao buscar parcelas", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(parcelas)
}

// POST /calculos-comissao/{cid}/parcelas
// POST /calculos-comissao/{cid}/parcelas
func (h *Handler) CreateForCalculo(w http.ResponseWriter, r *http.Request) {
	cid, err := strconv.Atoi(mux.Vars(r)["cid"])
	if err != nil {
		http.Error(w, "ID do cálculo inválido", http.StatusBadRequest)
		return
	}

	var in ParcelaCreateDTO
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "JSON mal formado", http.StatusBadRequest)
		return
	}
	if in.Status == "" {
		in.Status = "Pendente"
	}

	tx := h.Repo.DB.Begin()
	if tx.Error != nil {
		http.Error(w, "Falha ao iniciar transação", http.StatusInternalServerError)
		return
	}

	parcela := &ParcelaComissao{
		CalculoComissaoID: uint(cid),
		Valor:             in.Valor,
		DataVencimento:    in.DataVencimento,
		Status:            in.Status,
		Anexo:             in.Anexo,
		NotaFiscal:        in.NotaFiscal, // <— NOVO
		VolumeMensal:      in.VolumeMensal,
	}

	if err := tx.Create(parcela).Error; err != nil {
		_ = tx.Rollback()
		http.Error(w, "Erro ao criar parcela", http.StatusInternalServerError)
		return
	}

	if err := recalcTotalForCalculo(tx, uint(cid)); err != nil {
		_ = tx.Rollback()
		http.Error(w, "Erro ao recalcular total do cálculo", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit().Error; err != nil {
		_ = tx.Rollback()
		http.Error(w, "Erro ao confirmar transação", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(parcela)
}

// PATCH /parcelas/{pid}/status
// Permite: "Pendente", "NF Enviada", "Pago", "Cancelada".
// Regra: não permite rebaixar uma parcela já "Pago".
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(mux.Vars(r)["pid"])
	if err != nil {
		http.Error(w, "ID da parcela inválido", http.StatusBadRequest)
		return
	}

	var payload struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON mal formado", http.StatusBadRequest)
		return
	}

	allowed := map[string]bool{
		"Pendente":   true,
		"NF Enviada": true,
		"Pago":       true,
		"Cancelada":  true, // << novo
	}
	if !allowed[payload.Status] {
		http.Error(w, "Status inválido. Use 'Pendente', 'NF Enviada', 'Pago' ou 'Cancelada'.", http.StatusBadRequest)
		return
	}

	parcelaAtual, err := h.Repo.FindByID(uint(pid))
	if err != nil {
		http.Error(w, "Parcela não encontrada", http.StatusNotFound)
		return
	}
	if parcelaAtual.Status == "Pago" && payload.Status != "Pago" {
		http.Error(w, "Não é permitido alterar o status de uma parcela já paga", http.StatusBadRequest)
		return
	}

	// Repository: se status == "Pago" => seta data_pagamento = now; senão zera (NULL)
	if err := h.Repo.UpdateStatus(uint(pid), payload.Status, time.Now()); err != nil {
		http.Error(w, "Erro ao atualizar status da parcela", http.StatusInternalServerError)
		return
	}

	parcela, err := h.Repo.FindByID(uint(pid))
	if err != nil {
		http.Error(w, "Erro ao buscar parcela atualizada", http.StatusInternalServerError)
		return
	}

	// Mantido por consistência (não impacta total_receber por status)
	if err := recalcTotalForCalculo(h.Repo.DB, parcela.CalculoComissaoID); err != nil {
		http.Error(w, "Erro ao recalcular total do cálculo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(parcela)
}

// POST /parcelas/{pid}/anexo
func (h *Handler) UpdateAnexo(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(mux.Vars(r)["pid"])
	if err != nil {
		http.Error(w, "ID da parcela inválido", http.StatusBadRequest)
		return
	}

	var payload struct {
		Anexo string `json:"anexo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON mal formado", http.StatusBadRequest)
		return
	}

	if err := h.Repo.UpdateAnexo(uint(pid), payload.Anexo); err != nil {
		http.Error(w, "Erro ao atualizar anexo da parcela", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message":"Anexo atualizado com sucesso"}`))
}

// DELETE /parcelas/{pid}/anexo
func (h *Handler) DeleteAnexo(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(mux.Vars(r)["pid"])
	if err != nil {
		http.Error(w, "ID da parcela inválido", http.StatusBadRequest)
		return
	}

	if err := h.Repo.UpdateAnexo(uint(pid), ""); err != nil {
		http.Error(w, "Erro ao remover anexo da parcela", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message":"Anexo removido com sucesso"}`))
}

// PUT /parcelas/{pid}  (implementation kept later in this file)

// DELETE /parcelas/{pid}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(mux.Vars(r)["pid"])
	if err != nil {
		http.Error(w, "ID da parcela inválido", http.StatusBadRequest)
		return
	}

	parcela, err := h.Repo.FindByID(uint(pid))
	if err != nil {
		http.Error(w, "Parcela não encontrada", http.StatusNotFound)
		return
	}

	tx := h.Repo.DB.Begin()
	if tx.Error != nil {
		http.Error(w, "Falha ao iniciar transação", http.StatusInternalServerError)
		return
	}

	if err := tx.Delete(&ParcelaComissao{}, uint(pid)).Error; err != nil {
		_ = tx.Rollback()
		http.Error(w, "Erro ao deletar parcela", http.StatusInternalServerError)
		return
	}

	if err := recalcTotalForCalculo(tx, parcela.CalculoComissaoID); err != nil {
		_ = tx.Rollback()
		http.Error(w, "Erro ao recalcular total do cálculo", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit().Error; err != nil {
		_ = tx.Rollback()
		http.Error(w, "Erro ao confirmar transação", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PUT /parcelas/{pid}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(mux.Vars(r)["pid"])
	if err != nil {
		http.Error(w, "ID da parcela inválido", http.StatusBadRequest)
		return
	}

	parcelaExistente, err := h.Repo.FindByID(uint(pid))
	if err != nil {
		http.Error(w, "Parcela não encontrada", http.StatusNotFound)
		return
	}

	var payload ParcelaUpdateDTO
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON mal formado", http.StatusBadRequest)
		return
	}

	parcelaExistente.Valor = payload.Valor
	parcelaExistente.DataVencimento = payload.DataVencimento
	parcelaExistente.Status = payload.Status
	parcelaExistente.Anexo = payload.Anexo
	parcelaExistente.NotaFiscal = payload.NotaFiscal
	parcelaExistente.VolumeMensal = payload.VolumeMensal

	if payload.Status == "Pago" && parcelaExistente.DataPagamento == nil {
		now := time.Now()
		parcelaExistente.DataPagamento = &now
	}

	if err := h.Repo.Update(parcelaExistente); err != nil {
		http.Error(w, "Erro ao atualizar a parcela", http.StatusInternalServerError)
		return
	}

	// Recalcula total do cálculo
	if err := recalcTotalForCalculo(h.Repo.DB, parcelaExistente.CalculoComissaoID); err != nil {
		http.Error(w, "Erro ao recalcular total do cálculo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(parcelaExistente)
}

/* ============================== Nota Fiscal ============================== */

// POST /parcelas/{pid}/nota-fiscal
func (h *Handler) UpdateNotaFiscal(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(mux.Vars(r)["pid"])
	if err != nil {
		http.Error(w, "ID da parcela inválido", http.StatusBadRequest)
		return
	}

	var payload struct {
		NotaFiscal string `json:"notaFiscal"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON mal formado", http.StatusBadRequest)
		return
	}

	if err := h.Repo.UpdateNotaFiscal(uint(pid), payload.NotaFiscal); err != nil {
		http.Error(w, "Erro ao atualizar nota fiscal da parcela", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message":"Nota fiscal atualizada com sucesso"}`))
}

// DELETE /parcelas/{pid}/nota-fiscal
func (h *Handler) DeleteNotaFiscal(w http.ResponseWriter, r *http.Request) {
	pid, err := strconv.Atoi(mux.Vars(r)["pid"])
	if err != nil {
		http.Error(w, "ID da parcela inválido", http.StatusBadRequest)
		return
	}

	if err := h.Repo.UpdateNotaFiscal(uint(pid), ""); err != nil {
		http.Error(w, "Erro ao remover nota fiscal da parcela", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"message":"Nota fiscal removida com sucesso"}`))
}
