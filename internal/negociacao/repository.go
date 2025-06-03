package negociacao

import "gorm.io/gorm"

type Repository interface {
	Criar(db *gorm.DB, n *Negociacao) error
	ListarPorConsultor(db *gorm.DB, consultorID uint) ([]Negociacao, error)
	BuscarPorID(db *gorm.DB, id uint) (*Negociacao, error)
	ListarTodos(db *gorm.DB) ([]Negociacao, error)
	Atualizar(db *gorm.DB, id uint, dados *Negociacao) error
	Deletar(db *gorm.DB, id uint) error
	BuscarPorCNPJ(db *gorm.DB, cnpj string) (*Negociacao, error)

}

type repositoryImpl struct{}

func NewRepository() Repository {
	return &repositoryImpl{}
}

func (r *repositoryImpl) Criar(db *gorm.DB, n *Negociacao) error {
	return db.Create(n).Error
}

func (r *repositoryImpl) ListarPorConsultor(db *gorm.DB, consultorID uint) ([]Negociacao, error) {
	var lista []Negociacao
	err := db.Where("consultor_id = ?", consultorID).Find(&lista).Error
	return lista, err
}
func (r *repositoryImpl) AtualizarComentario(db *gorm.DB, id uint, comentario string) error {
	return db.Model(&Negociacao{}).
		Where("id = ?", id).
		Update("comentarios", comentario).Error
}
func (r *repositoryImpl) BuscarPorID(db *gorm.DB, id uint) (*Negociacao, error) {
	var n Negociacao
	err := db.First(&n, id).Error
	return &n, err
}

func (r *repositoryImpl) ListarTodos(db *gorm.DB) ([]Negociacao, error) {
	var lista []Negociacao
	err := db.Find(&lista).Error
	return lista, err
}

func (r *repositoryImpl) Atualizar(db *gorm.DB, id uint, dados *Negociacao) error {
	var atual Negociacao
	if err := db.First(&atual, id).Error; err != nil {
		return err
	}
	atual.Nome = dados.Nome
	atual.Sobrenome = dados.Sobrenome
	atual.Telefone = dados.Telefone
	atual.AnexoFatura = dados.AnexoFatura
	atual.AnexoEstudo = dados.AnexoEstudo
	atual.ContratoKC = dados.ContratoKC
	atual.AnexoContratoSocial = dados.AnexoContratoSocial
	atual.Status = dados.Status
	atual.Produtos = dados.Produtos
	atual.KromaTake = dados.KromaTake
	atual.UF = dados.UF
	return db.Save(&atual).Error
}

func (r *repositoryImpl) Deletar(db *gorm.DB, id uint) error {
	return db.Delete(&Negociacao{}, id).Error
}
func (r *repositoryImpl) BuscarPorCNPJ(db *gorm.DB, cnpj string) (*Negociacao, error) {
	var n Negociacao
	err := db.Where("cnpj = ?", cnpj).First(&n).Error
	if err != nil {
		return nil, err
	}
	return &n, nil
}