package consultor

import (
	"api/internal/contrato"
	"api/internal/negociacao"
	"time"
	"gorm.io/gorm"
)

type Repository interface {
	BuscarPorEmailOuCNPJ(db *gorm.DB, valor string) (*Consultor, error)
	Salvar(db *gorm.DB, c *Consultor) error
	BuscarPorID(db *gorm.DB, id uint) (*Consultor, error)
	ListarTodos(db *gorm.DB) ([]Consultor, error)
	Atualizar(db *gorm.DB, id uint, novosDados *Consultor) error
	Deletar(db *gorm.DB, id uint) error
}

type repositoryImpl struct{}

func NewRepository() Repository {
	return &repositoryImpl{}
}

func (r *repositoryImpl) BuscarPorEmailOuCNPJ(db *gorm.DB, valor string) (*Consultor, error) {
	var consultor Consultor
	err := db.Where("email = ? OR cnpj = ?", valor, valor).First(&consultor).Error
	return &consultor, err
}

func (r *repositoryImpl) Salvar(db *gorm.DB, c *Consultor) error {
	return db.Save(c).Error
}

func (r *repositoryImpl) BuscarPorID(db *gorm.DB, id uint) (*Consultor, error) {
	var consultor Consultor
	err := db.Preload("Negociacoes.Contrato").
		Preload("Negociacoes.Comentarios").
		Preload("Contratos").
		First(&consultor, id).Error
	return &consultor, err
}

func (r *repositoryImpl) ListarTodos(db *gorm.DB) ([]Consultor, error) {
	var consultores []Consultor
	err := db.Preload("Negociacoes.Contrato").
		Preload("Negociacoes.Comentarios").
		Preload("Contratos").
		Find(&consultores).Error
	return consultores, err
}

func (r *repositoryImpl) Atualizar(db *gorm.DB, id uint, novosDados *Consultor) error {
	var existente Consultor
	if err := db.First(&existente, id).Error; err != nil {
		return err
	}
	existente.Nome = novosDados.Nome
	existente.Sobrenome = novosDados.Sobrenome
	existente.CNPJ = novosDados.CNPJ
	existente.Email = novosDados.Email
	existente.Telefone = novosDados.Telefone
	existente.Foto = novosDados.Foto

	return db.Save(&existente).Error
}

func (r *repositoryImpl) Deletar(db *gorm.DB, id uint) error {
	return db.Delete(&Consultor{}, id).Error
}

// MontarResumoConsultorDTO gera um resumo de métricas do consultor
func MontarResumoConsultorDTO(
	consultor Consultor,
	contratos []contrato.Contrato,
	negociacoes []negociacao.Negociacao,
) ResumoConsultorDTO {
	var recebida, aReceber float64
	now := time.Now()

	for _, c := range contratos {
		if c.ValorIntegral {
			if !now.Before(c.InicioSuprimento) {
				recebida += c.Valor
			} else {
				aReceber += c.Valor
			}
		} else {
			if !now.Before(c.FimSuprimento) {
				recebida += c.Valor
			} else {
				aReceber += c.Valor
			}
		}
	}

	ativas := 0
	for _, n := range negociacoes {
		if n.Status == "ativa" {
			ativas++
		}
	}

	return ResumoConsultorDTO{
		ID:                consultor.ID,
		Nome:              consultor.Nome,
		Sobrenome:         consultor.Sobrenome,
		Email:             consultor.Email,
		CNPJ:              consultor.CNPJ,
		Telefone:          consultor.Telefone,
		Foto:              consultor.Foto,
		ContratosFechados: len(contratos),
		NegociacoesAtivas: ativas,
		ComissaoRecebida:  recebida,
		ComissaoAReceber:  aReceber,
	}
}