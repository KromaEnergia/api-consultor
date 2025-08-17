package comentario

import (
	"time"

	"github.com/KromaEnergia/api-consultor/internal/models"
)

type AuthorDTO struct {
	Type string `json:"type"`           // "consultor" | "comercial" | "system"
	ID   *uint  `json:"id,omitempty"`   // nil para system
	Nome string `json:"nome,omitempty"` // genérico (sem preload)
	Foto string `json:"foto,omitempty"` // vazio por enquanto
}

type CommentDTO struct {
	ID           uint      `json:"ID"`
	NegociacaoID uint      `json:"negociacaoId"`
	Texto        string    `json:"texto"`
	System       bool      `json:"system"`
	CreatedAt    time.Time `json:"CreatedAt"`
	Author       AuthorDTO `json:"author"`
}

func toDTO(c models.Comentario) CommentDTO {
	out := CommentDTO{
		ID:           c.ID,
		NegociacaoID: c.NegociacaoID,
		Texto:        c.Texto,
		System:       c.IsSystem,
		CreatedAt:    c.CreatedAt,
	}

	if c.IsSystem {
		out.Author = AuthorDTO{Type: "system", Nome: "Sistema"}
		return out
	}

	if c.ComercialID != nil {
		out.Author = AuthorDTO{
			Type: "comercial",
			ID:   c.ComercialID,
			Nome: "Comercial",
			Foto: "",
		}
		return out
	}

	if c.ConsultorID > 0 {
		id := c.ConsultorID
		out.Author = AuthorDTO{
			Type: "consultor",
			ID:   &id,
			Nome: "Consultor",
			Foto: "",
		}
		return out
	}

	// fallback defensivo
	out.Author = AuthorDTO{Type: "consultor", Nome: "Usuário"}
	return out
}

func toDTOs(list []models.Comentario) []CommentDTO {
	out := make([]CommentDTO, 0, len(list))
	for _, c := range list {
		out = append(out, toDTO(c))
	}
	return out
}
