// internal/comercial/dto.go
package comercial

// LoginRequest é usado em POST /comerciais/login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CreateComercialRequest é usado em POST /comerciais
type CreateComercialRequest struct {
	Nome      string `json:"nome"`
	Sobrenome string `json:"sobrenome"`
	Documento string `json:"documento"`
	Email     string `json:"email"`
	Telefone  string `json:"telefone"`
	Foto      string `json:"foto"`
	Senha     string `json:"senha"`
	IsAdmin   bool   `json:"isAdmin"`
}

// UpdateComercialRequest é usado em PUT /comerciais/{id}
// Campos como ponteiro permitem omitir no JSON se não quiser alterar
type UpdateComercialRequest struct {
	Nome      *string `json:"nome,omitempty"`
	Sobrenome *string `json:"sobrenome,omitempty"`
	Telefone  *string `json:"telefone,omitempty"`
	Foto      *string `json:"foto,omitempty"`
}
