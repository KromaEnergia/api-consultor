package contrato

import "gorm.io/gorm"

type Repository interface {
	Criar(db *gorm.DB, c *Contrato) error
	BuscarPorNegociacao(db *gorm.DB, negociacaoID uint) (*Contrato, error)
	ListarTodos(db *gorm.DB) ([]Contrato, error)
}

type repositoryImpl struct{}

func NewRepository() Repository {
	return &repositoryImpl{}
}

func (r *repositoryImpl) Criar(db *gorm.DB, c *Contrato) error {
	return db.Create(c).Error
}

func (r *repositoryImpl) BuscarPorNegociacao(db *gorm.DB, negociacaoID uint) (*Contrato, error) {
	var contrato Contrato
	err := db.Where("negociacao_id = ?", negociacaoID).First(&contrato).Error
	return &contrato, err
}

func (r *repositoryImpl) ListarTodos(db *gorm.DB) ([]Contrato, error) {
	var contratos []Contrato
	err := db.Find(&contratos).Error
	return contratos, err
}
