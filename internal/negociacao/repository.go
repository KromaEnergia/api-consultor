// internal/negociacao/repository.go
package negociacao

import (
	"gorm.io/gorm"
)

// Repository define operações de persistência para Negociacao
type Repository interface {
	Salvar(db *gorm.DB, n *Negociacao) error
	ListarPorConsultor(db *gorm.DB, consultorID uint) ([]Negociacao, error)
	BuscarPorID(db *gorm.DB, id uint) (*Negociacao, error)
	Atualizar(db *gorm.DB, n *Negociacao) error
	Deletar(db *gorm.DB, id uint) error
}

// repositoryImpl implementa Repository
type repositoryImpl struct{}

// NewRepository cria instância de Repository
func NewRepository() Repository {
	return &repositoryImpl{}
}

// Salvar cria nova negociação
func (r *repositoryImpl) Salvar(db *gorm.DB, n *Negociacao) error {
	return db.Create(n).Error
}

// ListarPorConsultor retorna negociações de um consultor, com associações carregadas
func (r *repositoryImpl) ListarPorConsultor(db *gorm.DB, consultorID uint) ([]Negociacao, error) {
	var list []Negociacao
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
func (r *repositoryImpl) BuscarPorID(db *gorm.DB, id uint) (*Negociacao, error) {
	var n Negociacao
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
func (r *repositoryImpl) Atualizar(db *gorm.DB, n *Negociacao) error {
	return db.Save(n).Error
}

// Deletar remove negociação do banco
func (r *repositoryImpl) Deletar(db *gorm.DB, id uint) error {
	return db.Delete(&Negociacao{}, id).Error
}
