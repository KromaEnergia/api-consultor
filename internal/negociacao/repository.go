package negociacao

import (
	"gorm.io/gorm"
)

type Repository interface {
	Salvar(db *gorm.DB, n *Negociacao) error
	ListarPorConsultor(db *gorm.DB, consultorID uint) ([]Negociacao, error)
	BuscarPorID(db *gorm.DB, id uint) (*Negociacao, error)
	Atualizar(db *gorm.DB, n *Negociacao) error
	Deletar(db *gorm.DB, id uint) error
}

type repositoryImpl struct{}

func NewRepository() Repository {
	return &repositoryImpl{}
}

func (r *repositoryImpl) Salvar(db *gorm.DB, n *Negociacao) error {
	return db.Create(n).Error
}

func (r *repositoryImpl) ListarPorConsultor(db *gorm.DB, consultorID uint) ([]Negociacao, error) {
	var list []Negociacao
	err := db.
		Where("consultor_id = ?", consultorID).
		Preload("Produtos").
		Preload("Contrato").
		Preload("Comentarios").
		Find(&list).Error
	return list, err
}

func (r *repositoryImpl) BuscarPorID(db *gorm.DB, id uint) (*Negociacao, error) {
	var n Negociacao
	err := db.
		Preload("Produtos").
		Preload("Contrato").
		Preload("Produtos").
		Preload("Comentarios").
		First(&n, id).Error
	return &n, err
}

func (r *repositoryImpl) Atualizar(db *gorm.DB, n *Negociacao) error {
	return db.Save(n).Error
}

func (r *repositoryImpl) Deletar(db *gorm.DB, id uint) error {
	return db.Delete(&Negociacao{}, id).Error
}
