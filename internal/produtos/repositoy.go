// internal/produtos/repository.go
package produtos

import "gorm.io/gorm"

// Repository encapsula operações de banco para produtos
type Repository struct {
	DB *gorm.DB
}

// NewRepository cria um novo repositório de produtos
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{DB: db}
}

// CreateMany insere múltiplos produtos
func (r *Repository) CreateMany(produtos []Produto) error {
	return r.DB.Create(&produtos).Error
}

// FindByNeg retorna todos os produtos de uma negociação específica
func (r *Repository) FindByNeg(idNeg uint) ([]Produto, error) {
	var ps []Produto
	err := r.DB.Where("negociacao_id = ?", idNeg).Find(&ps).Error
	return ps, err
}

// FindByID busca um produto pelo seu ID
func (r *Repository) FindByID(id uint) (*Produto, error) {
	var p Produto
	if err := r.DB.First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// Update salva alterações em um produto existente
func (r *Repository) Update(p *Produto) error {
	return r.DB.Save(p).Error
}

// Delete remove um produto
func (r *Repository) Delete(p *Produto) error {
	return r.DB.Delete(p).Error
}

// ListAll retorna todos os produtos
func (r *Repository) ListAll() ([]Produto, error) {
	var produtos []Produto
	err := r.DB.Find(&produtos).Error
	return produtos, err
}

// ListarPorConsultor retorna todos os produtos de todas as negociações de um consultor
func (r *Repository) ListarPorConsultor(consultorID uint) ([]Produto, error) {
	var ps []Produto
	err := r.DB.
		Joins("JOIN negociacaos n ON n.id = produtos.negociacao_id").
		Where("n.consultor_id = ?", consultorID).
		Find(&ps).Error
	return ps, err
}
