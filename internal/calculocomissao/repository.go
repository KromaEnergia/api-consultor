// internal/calculocomissao/repository.go
package calculocomissao

import (
	"gorm.io/gorm"
)

// Repository encapsula operações de banco para CalculoComissao
type Repository struct {
	DB *gorm.DB
}

// NewRepository cria um novo repositório
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

// Create insere um novo cálculo de comissão
func (r *Repository) Create(calc *CalculoComissao) error {
	return r.DB.Create(calc).Error
}

// FindByNegociacao retorna todos cálculos para uma negociação
func (r *Repository) FindByNegociacao(negID uint) ([]CalculoComissao, error) {
	var list []CalculoComissao
	err := r.DB.Where("negociacao_id = ?", negID).Find(&list).Error
	return list, err
}

// <<<---- NOVA FUNÇÃO ---->>>
// FindByStatus retorna todos os cálculos com um determinado status.
func (r *Repository) FindByStatus(status string) ([]CalculoComissao, error) {
	var list []CalculoComissao
	err := r.DB.Where("status = ?", status).Find(&list).Error
	return list, err
}

// FindByID retorna um cálculo pelo ID
func (r *Repository) FindByID(id uint) (*CalculoComissao, error) {
	var calc CalculoComissao
	if err := r.DB.First(&calc, id).Error; err != nil {
		return nil, err
	}
	return &calc, nil
}

// Update salva alterações em um cálculo existente (atualiza todos os campos)
func (r *Repository) Update(calc *CalculoComissao) error {
	return r.DB.Save(calc).Error
}

// <<<---- NOVA FUNÇÃO ---->>>
// UpdateStatusAndTotal atualiza apenas o status e o total a receber de um cálculo.
// É mais eficiente que o Update geral para esta operação específica.
func (r *Repository) UpdateStatusAndTotal(id uint, status string, totalReceber float64) error {
	return r.DB.Model(&CalculoComissao{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        status,
		"total_receber": totalReceber,
	}).Error
}

// Delete remove um cálculo do banco (soft delete se o model tiver gorm.DeletedAt)
func (r *Repository) Delete(calc *CalculoComissao) error {
	return r.DB.Delete(calc).Error
}

// NO ARQUIVO repository.go

// FindByNegociacaoAndStatus retorna todos os cálculos para uma negociação com um status específico.
func (r *Repository) FindByNegociacaoAndStatus(negID uint, status string) ([]CalculoComissao, error) {
	var list []CalculoComissao
	err := r.DB.Where("negociacao_id = ? AND status = ?", negID, status).Find(&list).Error
	return list, err
}

// ... outros métodos do repositório

// <<<---- NOVO MÉTODO ---->>>
// ListByConsultorID busca todos os cálculos de comissão de um consultor,
// pré-carregando as parcelas de cada um.
func (r *Repository) ListByConsultorID(consultorID uint) ([]CalculoComissao, error) {
	var calculos []CalculoComissao
	err := r.DB.
		// Pré-carrega as parcelas para cada cálculo encontrado
		Preload("Parcelas").
		// Junta com a tabela de negociações para poder filtrar pelo consultor
		Joins("JOIN negociacoes ON negociacoes.id = calculo_comissaos.negociacao_id").
		// Filtra pelo ID do consultor
		Where("negociacoes.consultor_id = ?", consultorID).
		// Executa a busca
		Find(&calculos).Error
	return calculos, err
}
