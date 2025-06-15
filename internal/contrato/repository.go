package contrato

import "gorm.io/gorm"

type Repository interface {
	Salvar(db *gorm.DB, c *Contrato) error
	BuscarPorNegociacao(db *gorm.DB, negID uint) (*Contrato, error)
	ListarPorConsultor(db *gorm.DB, consultorID uint) ([]Contrato, error)
	Atualizar(db *gorm.DB, c *Contrato) error
	Deletar(db *gorm.DB, id uint) error
}

type repositoryImpl struct{}

func NewRepository() Repository { return &repositoryImpl{} }

func (r *repositoryImpl) Salvar(db *gorm.DB, c *Contrato) error {
	return db.Create(c).Error
}

func (r *repositoryImpl) BuscarPorNegociacao(db *gorm.DB, negID uint) (*Contrato, error) {
	var c Contrato
	err := db.Where("negociacao_id = ?", negID).First(&c).Error
	return &c, err
}

func (r *repositoryImpl) ListarPorConsultor(db *gorm.DB, consultorID uint) ([]Contrato, error) {
	var list []Contrato
	err := db.Joins("JOIN negociacaos n ON n.id = contratos.negociacao_id").
		Where("n.consultor_id = ?", consultorID).
		Find(&list).Error
	return list, err
}

func (r *repositoryImpl) Atualizar(db *gorm.DB, c *Contrato) error {
	return db.Save(c).Error
}

func (r *repositoryImpl) Deletar(db *gorm.DB, id uint) error {
	return db.Delete(&Contrato{}, id).Error
}
