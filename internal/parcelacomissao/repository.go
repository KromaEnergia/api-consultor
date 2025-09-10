// internal/parcelacomissao/repository.go
package parcelacomissao

import (
	"time"

	"gorm.io/gorm"
)

// Repository encapsula o acesso a dados de Parcelas de Comissão.
type Repository struct {
	DB *gorm.DB
}

// NewRepository instancia um novo repositório.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

// WithDB retorna uma cópia do repo usando um *gorm.DB específico (ex.: tx).
func (r *Repository) WithDB(db *gorm.DB) *Repository {
	if db == nil {
		db = r.DB
	}
	return &Repository{DB: db}
}

/* ========================= CRUD básico de parcelas ========================= */

// CreateInBatch cria múltiplas parcelas de uma vez (ignora se vazio).
func (r *Repository) CreateInBatch(parcelas []*ParcelaComissao) error {
	if len(parcelas) == 0 {
		return nil
	}
	return r.DB.Create(parcelas).Error
}

// CreateForCalculo cria uma parcela vinculada a um cálculo específico.
func (r *Repository) CreateForCalculo(calculoID uint, p *ParcelaComissao) error {
	p.CalculoComissaoID = calculoID
	return r.DB.Create(p).Error
}

// FindByID busca uma única parcela pelo seu ID.
func (r *Repository) FindByID(id uint) (*ParcelaComissao, error) {
	var parcela ParcelaComissao
	if err := r.DB.First(&parcela, id).Error; err != nil {
		return nil, err
	}
	return &parcela, nil
}

// ListByCalculoID busca todas as parcelas de um cálculo de comissão específico.
func (r *Repository) ListByCalculoID(calculoID uint) ([]ParcelaComissao, error) {
	var parcelas []ParcelaComissao
	err := r.DB.
		Where("calculo_comissao_id = ?", calculoID).
		Order("data_vencimento ASC").
		Find(&parcelas).Error
	return parcelas, err
}

// ListByConsultorID busca todas as parcelas de todas as negociações de um consultor.
func (r *Repository) ListByConsultorID(consultorID uint) ([]ParcelaComissao, error) {
	var parcelas []ParcelaComissao
	err := r.DB.
		Table("parcela_comissaos").
		Select("parcela_comissaos.*").
		Joins("JOIN calculo_comissaos ON calculo_comissaos.id = parcela_comissaos.calculo_comissao_id").
		Joins("JOIN negociacoes ON negociacoes.id = calculo_comissaos.negociacao_id").
		Where("negociacoes.consultor_id = ?", consultorID).
		Order("data_vencimento ASC").
		Find(&parcelas).Error

	return parcelas, err
}

// Update atualiza todos os campos de uma parcela existente (Save exige PK).
func (r *Repository) Update(parcela *ParcelaComissao) error {
	return r.DB.Save(parcela).Error
}

// DeleteByID apaga a parcela; retorna gorm.ErrRecordNotFound se nada foi deletado.
func (r *Repository) DeleteByID(id uint) error {
	res := r.DB.Delete(&ParcelaComissao{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

/* ============================= Atualizações parciais ============================= */

// UpdateAnexo atualiza o campo 'anexo' de uma parcela.
func (r *Repository) UpdateAnexo(id uint, anexo string) error {
	return r.DB.Model(&ParcelaComissao{}).
		Where("id = ?", id).
		Update("anexo", anexo).Error
}

// UpdateStatus atualiza o status e ajusta data_pagamento.
// - Se status == "Pago", define data_pagamento = data informada.
// - Caso contrário, zera data_pagamento (NULL).
func (r *Repository) UpdateStatus(id uint, status string, dataPagamento time.Time) error {
	updates := map[string]interface{}{"status": status}
	if status == "Pago" {
		updates["data_pagamento"] = &dataPagamento
	} else {
		updates["data_pagamento"] = nil
	}
	return r.DB.Model(&ParcelaComissao{}).
		Where("id = ?", id).
		Updates(updates).Error
}

/* ======================= Soma e recálculo do total_receber ======================= */

// SumValorByCalculoID soma os valores das parcelas de um cálculo.
// Se db == nil, usa o r.DB. Permite usar dentro de transação.
func (r *Repository) SumValorByCalculoID(db *gorm.DB, calculoID uint) (float64, error) {
	if db == nil {
		db = r.DB
	}
	var total float64
	err := db.Model(&ParcelaComissao{}).
		Where("calculo_comissao_id = ?", calculoID).
		Select("COALESCE(SUM(valor), 0)").
		Scan(&total).Error
	return total, err
}

// RecalcTotalForCalculo calcula a soma e atualiza calculo_comissaos.total_receber.
// Se db == nil, usa o r.DB. Permite usar dentro de transação.
func (r *Repository) RecalcTotalForCalculo(db *gorm.DB, calculoID uint) error {
	if db == nil {
		db = r.DB
	}
	total, err := r.SumValorByCalculoID(db, calculoID)
	if err != nil {
		return err
	}
	return db.Table("calculo_comissaos").
		Where("id = ?", calculoID).
		Update("total_receber", total).Error
}
