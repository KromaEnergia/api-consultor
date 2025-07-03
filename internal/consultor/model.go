// internal/consultor/model.go
package consultor

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/KromaEnergia/api-consultor/internal/contrato"
	"github.com/KromaEnergia/api-consultor/internal/negociacao"
	"gorm.io/gorm"
)

// --- CustomDate para formato dd/mm/ano ---

const customDateLayout = "02/01/2006"

type CustomDate struct {
	time.Time
}

// ... (funções do CustomDate permanecem as mesmas) ...
func (cd *CustomDate) UnmarshalJSON(b []byte) (err error) {
	s := string(b)
	if s == "null" || s == `""` {
		return nil
	}
	s = s[1 : len(s)-1]
	t, err := time.Parse(customDateLayout, s)
	if err != nil {
		return err
	}
	cd.Time = t
	return nil
}
func (cd CustomDate) MarshalJSON() ([]byte, error) {
	if cd.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, cd.Time.Format(customDateLayout))), nil
}
func (cd CustomDate) Value() (driver.Value, error) { return cd.Time, nil }
func (cd *CustomDate) Scan(value interface{}) error {
	if value == nil {
		cd.Time = time.Time{}
		return nil
	}
	t, ok := value.(time.Time)
	if !ok {
		return fmt.Errorf("failed to cast value to time.Time")
	}
	cd.Time = t
	return nil
}

// --- Struct Consultor Atualizada ---

type Consultor struct {
	gorm.Model
	Nome                  string                  `json:"nome"`
	Sobrenome             string                  `json:"sobrenome"`
	CNPJ                  string                  `json:"cnpj" gorm:"unique"`
	RequestedCNPJ         string                  `json:"requestedCnpj,omitempty"`
	CNPJChangeApproved    bool                    `json:"cnpjChangeApproved,omitempty"`
	Email                 string                  `json:"email" gorm:"unique"`
	RequestedEmail        string                  `json:"requestedEmail,omitempty"`
	EmailChangeApproved   bool                    `json:"emailChangeApproved,omitempty"`
	Telefone              string                  `json:"telefone"`
	Foto                  string                  `json:"foto"`
	DataNascimento        CustomDate              `json:"dataNascimento,omitempty"`
	Estado                string                  `json:"estado,omitempty"`
	TermoDeParceria       string                  `json:"termoDeParceria"`
	Senha                 string                  `json:"-"`
	PrecisaRedefinirSenha bool                    `json:"-"`
	IsAdmin               bool                    `json:"isAdmin"`
	ComercialID           uint                    `gorm:"not null" json:"comercial_id"`
	Negociacoes           []negociacao.Negociacao `gorm:"foreignKey:ConsultorID" json:"negociacoes"`
	Contratos             []contrato.Contrato     `gorm:"foreignKey:ConsultorID" json:"contratos"`
}
