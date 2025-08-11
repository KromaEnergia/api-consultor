// internal/consultor/repository.go
package consultor

import (
	"strings"
	"time"

	"github.com/KromaEnergia/api-consultor/internal/contrato"
	"github.com/KromaEnergia/api-consultor/internal/models"
	"github.com/KromaEnergia/api-consultor/internal/produtos"
	"gorm.io/gorm"
)

// Repository define a interface para as operações de banco de dados do consultor.
type Repository interface {
	BuscarPorEmail(db *gorm.DB, valor string) (*Consultor, error)
	Salvar(db *gorm.DB, c *Consultor) error
	BuscarPorID(db *gorm.DB, id uint) (*Consultor, error)
	ListarTodos(db *gorm.DB) ([]Consultor, error)
	Atualizar(db *gorm.DB, id uint, novosDados *Consultor) error
	Deletar(db *gorm.DB, id uint) error
	ListarTodosSimples(db *gorm.DB) ([]Consultor, error)
	ListarComPreload(db *gorm.DB) ([]Consultor, error)
	// Métodos para Dados Bancários
	GetDadosBancarios(db *gorm.DB, consultorID uint) (DadosBancarios, error)
	UpdateDadosBancarios(db *gorm.DB, consultorID uint, dados DadosBancarios) error
	DeleteDadosBancarios(db *gorm.DB, consultorID uint) error
}

type repositoryImpl struct{}

// NewRepository cria uma nova instância do repositório.
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
		First(&c, id).Error
	return &c, err
}

func (r *repositoryImpl) ListarTodos(db *gorm.DB) ([]Consultor, error) {
	var consultores []Consultor
	err := db.Preload("Negociacoes.Contrato").
		Preload("Negociacoes.Comentarios").
		Preload("Contratos"). // Corrigido de "Contrato" para "Contratos"
		Preload("Negociacoes.CalculoComissao").
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

func (r *repositoryImpl) ListarTodosSimples(db *gorm.DB) ([]Consultor, error) {
	var consultores []Consultor
	err := db.Find(&consultores).Error
	return consultores, err
}

func (r *repositoryImpl) ListarComPreload(db *gorm.DB) ([]Consultor, error) {
	var consultores []Consultor
	err := db.
		Preload("Negociacoes").
		Preload("Negociacoes.Produtos").
		Preload("Negociacoes.Comentarios").
		Preload("Negociacoes.CalculosComissao").
		Preload("Negociacoes.CalculosComissao.Parcelas").
		Preload("Contratos").
		Find(&consultores).Error
	return consultores, err
}

// --- Implementação dos Métodos para Dados Bancários ---

// GetDadosBancarios busca os dados bancários de um consultor específico.
func (r *repositoryImpl) GetDadosBancarios(db *gorm.DB, consultorID uint) (DadosBancarios, error) {
	var consultor Consultor
	result := db.First(&consultor, consultorID)
	if result.Error != nil {
		return DadosBancarios{}, result.Error
	}
	return consultor.DadosBancarios, nil
}

// UpdateDadosBancarios atualiza ou cria os dados bancários de um consultor.
func (r *repositoryImpl) UpdateDadosBancarios(db *gorm.DB, consultorID uint, dados DadosBancarios) error {
	// O GORM lida com a serialização do struct para JSON automaticamente
	// graças aos métodos Value() e Scan() que implementamos no model.
	result := db.Model(&Consultor{}).Where("id = ?", consultorID).Update("dados_bancarios", dados)
	return result.Error
}

// DeleteDadosBancarios remove os dados bancários de um consultor.
func (r *repositoryImpl) DeleteDadosBancarios(db *gorm.DB, consultorID uint) error {
	// Define o campo como um objeto JSON vazio.
	// Isso é mais seguro do que setar para NULL, dependendo da sua regra de negócio.
	dadosVazios := DadosBancarios{}
	result := db.Model(&Consultor{}).Where("id = ?", consultorID).Update("dados_bancarios", dadosVazios)
	return result.Error
}

// Monta um DTO com os principais dados e métricas do consultor
func MontarResumoConsultorDTO(
	consultor Consultor,
	contratos []contrato.Contrato,
	negociacoes []models.Negociacao,
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
		Produtos:          tipos, // preenche o novo campo
	}
}
