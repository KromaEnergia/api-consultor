package main

import (
	"api/internal/comentario"
	"api/internal/consultor"
	"api/internal/contrato"
	"api/internal/negociacao"
	"api/internal/notificacao"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=sistema port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Erro ao conectar no banco:", err)
	}

	// AutoMigrate para todos os modelos
	if err := db.AutoMigrate(
		&consultor.Consultor{},
		&negociacao.Negociacao{},
		&comentario.Comentario{},
		&contrato.Contrato{},
	); err != nil {
		log.Fatal("Erro no AutoMigrate:", err)
	}

	// Handlers
	consultorHandler := consultor.NewHandler(db)
	negociacaoHandler := negociacao.NewHandler(db)
	comentarioHandler := comentario.NewHandler(db)
	contratoHandler := contrato.NewHandler(db)
	notificacaoHandler := notificacao.NewHandler()

	// Router
	r := mux.NewRouter()

	// Rotas de consultores
	r.HandleFunc("/consultores", consultorHandler.CriarConsultor).Methods("POST")
	r.HandleFunc("/consultores", consultorHandler.ListarConsultores).Methods("GET")
	r.HandleFunc("/consultores/{id}", consultorHandler.BuscarPorID).Methods("GET")
	r.HandleFunc("/consultores/{id}", consultorHandler.AtualizarConsultor).Methods("PUT")
	r.HandleFunc("/consultores/{id}", consultorHandler.DeletarConsultor).Methods("DELETE")

	// Rotas de negociações
	r.HandleFunc("/negociacoes", negociacaoHandler.CriarNegociacao).Methods("POST")
	r.HandleFunc("/negociacoes", negociacaoHandler.ListarNegociacoes).Methods("GET")
	r.HandleFunc("/negociacoes/{id}", negociacaoHandler.BuscarPorID).Methods("GET")
	r.HandleFunc("/negociacoes/{id}", negociacaoHandler.AtualizarNegociacao).Methods("PUT")
	r.HandleFunc("/negociacoes/{id}", negociacaoHandler.DeletarNegociacao).Methods("DELETE")

	// Rotas de contratos
	r.HandleFunc("/contratos", contratoHandler.CriarContrato).Methods("POST")
	r.HandleFunc("/contratos", contratoHandler.ListarContratos).Methods("GET")
	r.HandleFunc("/contratos/{id}", contratoHandler.BuscarPorID).Methods("GET")
	r.HandleFunc("/contratos/{id}", contratoHandler.AtualizarContrato).Methods("PUT")
	r.HandleFunc("/contratos/{id}", contratoHandler.DeletarContrato).Methods("DELETE")

	// Rotas de comentários
	r.HandleFunc("/comentarios", comentarioHandler.CriarComentario).Methods("POST")
	r.HandleFunc("/comentarios", comentarioHandler.ListarTodos).Methods("GET")
	r.HandleFunc("/comentarios/{id}", comentarioHandler.BuscarPorID).Methods("GET")
	r.HandleFunc("/comentarios/{id}", comentarioHandler.Atualizar).Methods("PUT")
	r.HandleFunc("/comentarios/{id}", comentarioHandler.RemoverComentario).Methods("DELETE")
	r.HandleFunc("/negociacoes/{id}/comentarios", comentarioHandler.ListarPorNegociacao).Methods("GET")

	// Rota de notificação (exemplo de alerta por CNPJ duplicado)
	r.HandleFunc("/notificar", notificacaoHandler.EnviarAlerta).Methods("POST")

	// Inicia servidor
	fmt.Println("Servidor rodando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
