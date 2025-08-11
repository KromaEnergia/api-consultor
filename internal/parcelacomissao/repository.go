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

// CreateInBatch cria múltiplas parcelas de uma vez.
func (r *Repository) CreateInBatch(parcelas []*ParcelaComissao) error {
	return r.DB.Create(parcelas).Error
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

// UpdateAnexo atualiza o campo de anexo de uma parcela específica.
func (r *Repository) UpdateAnexo(id uint, anexo string) error {
	result := r.DB.Model(&ParcelaComissao{}).Where("id = ?", id).Update("anexo", anexo)
	return result.Error
}

// FindByID busca uma única parcela pelo seu ID.
func (r *Repository) FindByID(id uint) (*ParcelaComissao, error) {
	var parcela ParcelaComissao
	if err := r.DB.First(&parcela, id).Error; err != nil {
		return nil, err
	}
	return &parcela, nil
}

// UpdateStatus atualiza o status e a data de pagamento de uma parcela.
func (r *Repository) UpdateStatus(id uint, status string, dataPagamento time.Time) error {
	updates := map[string]interface{}{"status": status}
	if status == "Pago" {
		updates["data_pagamento"] = &dataPagamento
	}
	return r.DB.Model(&ParcelaComissao{}).Where("id = ?", id).Updates(updates).Error
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

// Update atualiza todos os campos de uma parcela existente.
func (r *Repository) Update(parcela *ParcelaComissao) error {
	// O método Save do GORM atualiza todos os campos do objeto se a chave primária existir.
	return r.DB.Save(parcela).Error
}
