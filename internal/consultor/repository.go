package consultor

import (
	"strings"
	"time"

	"github.com/KromaEnergia/api-consultor/internal/contrato"
	"github.com/KromaEnergia/api-consultor/internal/negociacao"
	"github.com/KromaEnergia/api-consultor/internal/produtos"
	"gorm.io/gorm"
)

type Repository interface {
	BuscarPorEmail(db *gorm.DB, valor string) (*Consultor, error)
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
func (r *repositoryImpl) BuscarPorEmail(db *gorm.DB, email string) (*Consultor, error) {
	var c Consultor
	err := db.
		Where("email = ?", email).
		First(&c).
		Error
	return &c, err
}

func (r *repositoryImpl) Salvar(db *gorm.DB, c *Consultor) error {
	return db.Save(c).Error
}

func (r *repositoryImpl) BuscarPorID(db *gorm.DB, id uint) (*Consultor, error) {
	var c Consultor
	err := db.
		Preload("Negociacoes").
		Preload("Negociacoes.Produtos").
		Preload("Negociacoes.Comentarios").
		Preload("Negociacoes.Contrato").
		Preload("Negociacoes.contrato").
		First(&c, id).Error
	return &c, err
}

func (r *repositoryImpl) ListarTodos(db *gorm.DB) ([]Consultor, error) {
	var consultores []Consultor
	err := db.Preload("Negociacoes.Contrato").
		Preload("Negociacoes.Comentarios").
		Preload("Contrato").
		Preload("Negociacoes.contrato").
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

// Monta um DTO com os principais dados e métricas do consultor
func MontarResumoConsultorDTO(
	consultor Consultor,
	contratos []contrato.Contrato,
	negociacoes []negociacao.Negociacao,
	produtos []produtos.Produto, // ← novo parâmetro
) ResumoConsultorDTO {
	var recebida, aReceber float64
	now := time.Now()

	// cálculo de comissões…
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

	// conta negociações ativas…
	ativas := 0
	for _, n := range negociacoes {
		statusLower := strings.ToLower(strings.TrimSpace(n.Status))
		if statusLower == "negociação ativa" || statusLower == "ativa" {
			ativas++
		}
	}

	// extrai apenas o tipo de cada produto
	tipos := make([]string, len(produtos))
	for i, p := range produtos {
		tipos[i] = p.Tipo
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

		Produtos: tipos, // preenche o novo campo
	}
}
