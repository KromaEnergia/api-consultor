// internal/negociacao/repository.go
package negociacao

import (
	"github.com/KromaEnergia/api-consultor/internal/models"
	"gorm.io/gorm"
)

// Repository define operações de persistência para models.Negociacao
type Repository interface {
	Salvar(db *gorm.DB, n *models.Negociacao) error
	ListarPorConsultor(db *gorm.DB, consultorID uint) ([]models.Negociacao, error)
	BuscarPorID(db *gorm.DB, id uint) (*models.Negociacao, error)
	Atualizar(db *gorm.DB, n *models.Negociacao) error
	Deletar(db *gorm.DB, id uint) error
	AtualizarStatus(db *gorm.DB, id uint, status string) error

	// AtualizarStatus(db *gorm.DB, id uint, status string) error
}

// repositoryImpl implementa Repository
type repositoryImpl struct{}

// NewRepository cria instância de Repository
func NewRepository() Repository {
	return &repositoryImpl{}
}

// Salvar cria nova negociação
func (r *repositoryImpl) Salvar(db *gorm.DB, n *models.Negociacao) error {
	return db.Create(n).Error
}

// ListarPorConsultor retorna negociações de um consultor, com associações carregadas
func (r *repositoryImpl) ListarPorConsultor(db *gorm.DB, consultorID uint) ([]models.Negociacao, error) {
	var list []models.Negociacao
	err := db.
		Where("consultor_id = ?", consultorID).
		Preload("Contratos").
		Preload("Produtos").
		Preload("Comentarios").
		Preload("CalculosComissao").
		Preload("CalculosComissao.Parcelas").
		Find(&list).Error
	return list, err
}

// BuscarPorID retorna uma negociação por ID, com associações carregadas
func (r *repositoryImpl) BuscarPorID(db *gorm.DB, id uint) (*models.Negociacao, error) {
	var n models.Negociacao
	err := db.
		Preload("Contratos").
		Preload("Produtos").
		Preload("Comentarios").
		Preload("CalculosComissao").
		Preload("CalculosComissao.Parcelas").
		First(&n, id).Error
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// Atualizar salva alterações em uma negociação existente
func (r *repositoryImpl) Atualizar(db *gorm.DB, n *models.Negociacao) error {
	return db.Save(n).Error
}

// Deletar remove negociação do banco
func (r *repositoryImpl) Deletar(db *gorm.DB, id uint) error {
	return db.Delete(&models.Negociacao{}, id).Error
}

// AtualizarStatus atualiza apenas o campo de status de uma negociação específica.
func (r *repositoryImpl) AtualizarStatus(db *gorm.DB, id uint, status string) error {
	// A operação de atualização é feita aqui
	result := db.Model(&models.Negociacao{}).Where("id = ?", id).Update("status", status)

	// O erro da operação é verificado primeiro
	if result.Error != nil {
		return result.Error
	}

	// MELHORIA: Verificamos se alguma linha foi afetada.
	// Se RowsAffected for 0, significa que o ID não foi encontrado.
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound // Retorna um erro padrão de "registro não encontrado"
	}

	return nil
}
