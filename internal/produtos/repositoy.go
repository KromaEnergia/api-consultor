// internal/produto/repository.go
package produto

import "gorm.io/gorm"

type Repository struct {
	DB *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) CreateMany(produtos []Produto) error {
	return r.DB.Create(&produtos).Error
}

func (r *Repository) FindByNeg(idNeg uint) ([]Produto, error) {
	var ps []Produto
	err := r.DB.Where("negociacao_id = ?", idNeg).Find(&ps).Error
	return ps, err
}

func (r *Repository) FindByID(id uint) (*Produto, error) {
	var p Produto
	if err := r.DB.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) Update(p *Produto) error {
	return r.DB.Save(p).Error
}

func (r *Repository) Delete(p *Produto) error {
	return r.DB.Delete(p).Error
}
func (r *Repository) ListAll() ([]Produto, error) {
	var produtos []Produto
	err := r.DB.Find(&produtos).Error
	return produtos, err
}
