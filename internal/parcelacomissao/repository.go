package parcelacomissao

import (
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

// CreateInBatch cria múltiplas parcelas de uma vez.
func (r *Repository) CreateInBatch(parcelas []*ParcelaComissao) error {
	return r.DB.Create(&parcelas).Error
}

// ListByCalculoID busca todas as parcelas de um cálculo de comissão específico.
func (r *Repository) ListByCalculoID(calculoID uint) ([]ParcelaComissao, error) {
	var parcelas []ParcelaComissao
	err := r.DB.Where("calculo_comissao_id = ?", calculoID).Order("data_vencimento ASC").Find(&parcelas).Error
	return parcelas, err
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
	return r.DB.Model(&ParcelaComissao{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":         status,
		"data_pagamento": dataPagamento,
	}).Error
}

// <<<---- NOVO MÉTODO ---->>>
// ListByConsultorID busca todas as parcelas de todas as negociações de um consultor.
func (r *Repository) ListByConsultorID(consultorID uint) ([]ParcelaComissao, error) {
	var parcelas []ParcelaComissao
	err := r.DB.
		// Seleciona todos os campos da tabela 'parcela_comissaos'
		Table("parcela_comissaos").
		Select("parcela_comissaos.*").
		// Junta com a tabela de cálculos de comissão
		Joins("JOIN calculo_comissaos ON calculo_comissaos.id = parcela_comissaos.calculo_comissao_id").
		// Junta com a tabela de negociações
		Joins("JOIN negociacoes ON negociacoes.id = calculo_comissaos.negociacao_id").
		// Filtra pelo ID do consultor na tabela de negociações
		Where("negociacoes.consultor_id = ?", consultorID).
		// Ordena o resultado pela data de vencimento
		Order("data_vencimento ASC").
		// Encontra os resultados e os coloca na variável 'parcelas'
		Find(&parcelas).Error

	return parcelas, err
}
