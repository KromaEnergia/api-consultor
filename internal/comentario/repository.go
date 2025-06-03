package comentario

import "gorm.io/gorm"

type Repository interface {
	Criar(db *gorm.DB, c *Comentario) error
	ListarPorNegociacao(db *gorm.DB, negociacaoID uint) ([]Comentario, error)
	Remover(db *gorm.DB, id uint) error
		ListarTodos(db *gorm.DB) ([]Comentario, error)
	BuscarPorID(db *gorm.DB, id uint) (*Comentario, error)
	Atualizar(db *gorm.DB, id uint, novoTexto string) error

}

type repositoryImpl struct{}

func NewRepository() Repository {
	return &repositoryImpl{}
}

func (r *repositoryImpl) Criar(db *gorm.DB, c *Comentario) error {
	return db.Create(c).Error
}

func (r *repositoryImpl) ListarPorNegociacao(db *gorm.DB, negociacaoID uint) ([]Comentario, error) {
	var comentarios []Comentario
	err := db.Where("negociacao_id = ?", negociacaoID).Find(&comentarios).Error
	return comentarios, err
}

func (r *repositoryImpl) Remover(db *gorm.DB, id uint) error {
	return db.Delete(&Comentario{}, id).Error
}
func (r *repositoryImpl) ListarTodos(db *gorm.DB) ([]Comentario, error) {
	var comentarios []Comentario
	err := db.Find(&comentarios).Error
	return comentarios, err
}

func (r *repositoryImpl) BuscarPorID(db *gorm.DB, id uint) (*Comentario, error) {
	var c Comentario
	err := db.First(&c, id).Error
	return &c, err
}

func (r *repositoryImpl) Atualizar(db *gorm.DB, id uint, novoTexto string) error {
	return db.Model(&Comentario{}).Where("id = ?", id).Update("texto", novoTexto).Error
}
