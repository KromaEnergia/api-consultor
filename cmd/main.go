package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"api/internal/comentario"
	"api/internal/consultor"
	"api/internal/contrato"
	"api/internal/negociacao"

	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Obtém DSN a partir da variável de ambiente ou usa fallback
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "host=db user=postgres password=postgres dbname=consultor port=5432 sslmode=disable"
	}

	// Conecta ao banco
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Erro ao conectar no banco: ", err)
	}

	// AutoMigrate todos os modelos
	if err := db.AutoMigrate(
		&consultor.Consultor{},
		&negociacao.Negociacao{},
		&comentario.Comentario{},
		&contrato.Contrato{},
	); err != nil {
		log.Fatal("Erro no AutoMigrate: ", err)
	}

	// Inicializa handlers
	consultorHandler := consultor.NewHandler(db)
	negHandler := negociacao.NewHandler(db)
	comentHandler := comentario.NewHandler(db)
	contratoHandler := contrato.NewHandler(db)

	// Configura roteador
	r := mux.NewRouter()

	// Rotas de consultores
	r.HandleFunc("/consultores", consultorHandler.CriarConsultor).Methods("POST")
	r.HandleFunc("/consultores", consultorHandler.ListarConsultores).Methods("GET")
	r.HandleFunc("/consultores/{id}", consultorHandler.BuscarPorID).Methods("GET")
	r.HandleFunc("/consultores/{id}", consultorHandler.AtualizarConsultor).Methods("PUT")
	r.HandleFunc("/consultores/{id}", consultorHandler.DeletarConsultor).Methods("DELETE")
	r.HandleFunc("/consultores/{id}/resumo", consultorHandler.ObterResumoConsultor).Methods("GET")

	// Rotas de negociações
	r.HandleFunc("/negociacoes", negHandler.Criar).Methods("POST")
	r.HandleFunc("/negociacoes/{id}", negHandler.BuscarPorID).Methods("GET")
	r.HandleFunc("/consultores/{id}/negociacoes", negHandler.ListarPorConsultor).Methods("GET")
	r.HandleFunc("/negociacoes/{id}", negHandler.Atualizar).Methods("PUT")
	r.HandleFunc("/negociacoes/{id}", negHandler.Deletar).Methods("DELETE")

	// Rotas de contratos
	r.HandleFunc("/negociacoes/{id}/contrato", contratoHandler.CriarParaNegociacao).Methods("POST")
	r.HandleFunc("/negociacoes/{id}/contrato", contratoHandler.BuscarPorNegociacao).Methods("GET")
	r.HandleFunc("/consultores/{id}/contratos", contratoHandler.ListarPorConsultor).Methods("GET")
	r.HandleFunc("/contratos/{id}", contratoHandler.Atualizar).Methods("PUT")
	r.HandleFunc("/contratos/{id}", contratoHandler.Deletar).Methods("DELETE")

	// Rotas de comentários
	r.HandleFunc("/negociacoes/{id}/comentarios", comentHandler.ListarPorNegociacao).Methods("GET")
	r.HandleFunc("/negociacoes/{id}/comentarios", comentHandler.CriarComentario).Methods("POST")
	r.HandleFunc("/comentarios/{id}", comentHandler.BuscarPorID).Methods("GET")
	r.HandleFunc("/comentarios/{id}", comentHandler.Atualizar).Methods("PUT")
	r.HandleFunc("/comentarios/{id}", comentHandler.RemoverComentario).Methods("DELETE")

	// Inicia servidor
	fmt.Println("Servidor rodando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
