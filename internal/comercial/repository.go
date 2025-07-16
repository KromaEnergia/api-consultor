package comercial

import (
	"gorm.io/gorm"
)

type Repository interface {
	FindByEmail(db *gorm.DB, email string) (*Comercial, error)
	Save(db *gorm.DB, c *Comercial) error
	ListAll(db *gorm.DB) ([]Comercial, error)
	FindByID(db *gorm.DB, id uint) (*Comercial, error)
	Update(db *gorm.DB, id uint, req *UpdateComercialRequest) error
	Delete(db *gorm.DB, id uint) error
}

type repositoryImpl struct{}

func NewRepository() Repository {
	return &repositoryImpl{}
}

func (r *repositoryImpl) FindByEmail(db *gorm.DB, email string) (*Comercial, error) {
	var c Comercial
	if err := db.Where("email = ?", email).First(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *repositoryImpl) Save(db *gorm.DB, c *Comercial) error {
	return db.Create(c).Error
}

func (r *repositoryImpl) ListAll(db *gorm.DB) ([]Comercial, error) {
	var list []Comercial
	err := db.
		Preload("Consultores", func(db *gorm.DB) *gorm.DB {
			return db.
				Preload("Negociacoes").
				Preload("Contratos").
				Preload("Negociacoes.Contratos").
				Preload("Negociacoes.Produtos")
		}).
		Find(&list).Error
	return list, err
}

func (r *repositoryImpl) FindByID(db *gorm.DB, id uint) (*Comercial, error) {
	var c Comercial
	err := db.
		Preload("Consultores", func(db *gorm.DB) *gorm.DB {
			return db.
				Preload("Negociacoes").
				Preload("Contratos").
				Preload("Negociacoes.Contratos").
				Preload("Negociacoes.Produtos")
		}).
		First(&c, id).Error
	return &c, err
}

func (r *repositoryImpl) Update(db *gorm.DB, id uint, req *UpdateComercialRequest) error {
	var c Comercial
	if err := db.First(&c, id).Error; err != nil {
		return err
	}
	if req.Nome != nil {
		c.Nome = *req.Nome
	}
	if req.Sobrenome != nil {
		c.Sobrenome = *req.Sobrenome
	}
	if req.Telefone != nil {
		c.Telefone = *req.Telefone
	}
	if req.Foto != nil {
		c.Foto = *req.Foto
	}
	return db.Save(&c).Error
}

func (r *repositoryImpl) Delete(db *gorm.DB, id uint) error {
	return db.Delete(&Comercial{}, id).Error
}
