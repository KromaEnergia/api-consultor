// internal/negociacao/model.go
package negociacao

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/KromaEnergia/api-consultor/internal/comentario"
	"github.com/KromaEnergia/api-consultor/internal/contrato"
	"gorm.io/gorm"
)

// --- Tipo Customizado para Slice de Strings ---

// StringSlice define um tipo para um slice de strings que o GORM
// pode salvar corretamente no banco de dados.
type StringSlice []string

// Value implementa a interface driver.Valuer.
// Converte o slice de strings em uma string JSON para ser salva no banco.
func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil // Salva como um array JSON vazio se for nulo
	}
	return json.Marshal(s)
}

// Scan implementa a interface sql.Scanner.
// Converte a string JSON do banco de dados em um slice de strings.
func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("tipo não suportado para StringSlice")
	}

	return json.Unmarshal(bytes, s)
}

// --- Struct Negociacao Atualizada ---

// Negociacao representa uma oportunidade de negócio de um consultor
type Negociacao struct {
	ID        uint           `gorm:"primaryKey" json:"negociacaoId"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`

	Nome                string      `json:"nome"`
	Contato             string      `json:"contato"`
	Telefone            string      `json:"telefone"`
	CNPJ                string      `json:"cnpj"`
	Logo                string      `json:"logo"`
	AnexoFatura         string      `json:"anexoFatura"`
	AnexoEstudo         string      `json:"anexoEstudo"`
	ContratoKC          string      `json:"contratoKC"`
	AnexoContratoSocial string      `json:"anexoContratoSocial"`
	Arquivos            StringSlice `gorm:"type:text" json:"arquivos,omitempty"`
	Status              string      `json:"status"`
	Produtos            string      `json:"produtos"`
	KromaTake           bool        `json:"kromaTake"`
	UF                  string      `json:"uf"`
	ConsultorID         uint        `json:"consultorId"`

	// Relação 1-1 com Contrato
	Contrato contrato.Contrato `gorm:"foreignKey:NegociacaoID" json:"contrato"`
	// Relação 1-N com Comentarios
	Comentarios []comentario.Comentario `gorm:"foreignKey:NegociacaoID" json:"comentarios"`
}
