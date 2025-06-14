package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/KromaEnergia/api-consultor/internal/auth"
	"github.com/KromaEnergia/api-consultor/internal/comentario"
	"github.com/KromaEnergia/api-consultor/internal/consultor"
	"github.com/KromaEnergia/api-consultor/internal/contrato"
	"github.com/KromaEnergia/api-consultor/internal/negociacao"

	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Conexão ao banco (igual a antes)…
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "host=db user=postgres password=postgres dbname=consultor port=5432 sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Erro ao conectar no banco: ", err)
	}
	if err := db.AutoMigrate(
		&consultor.Consultor{},
		&negociacao.Negociacao{},
		&comentario.Comentario{},
		&contrato.Contrato{},
	); err != nil {
		log.Fatal("Erro no AutoMigrate: ", err)
	}

	r := mux.NewRouter()
	consultorHandler := consultor.NewHandler(db)

	// ─── Rotas PÚBLICAS (sem JWT) ─────────────────────────────
	r.HandleFunc("/login", consultorHandler.Login).Methods("POST")
	r.HandleFunc("/consultores", consultorHandler.CriarConsultor).Methods("POST")

	// ─── Subrouter PROTEGIDO (com JWT) ────────────────────────
	api := r.NewRoute().Subrouter()
	api.Use(auth.MiddlewareAutenticacao)

	// Consultores (exceto criação)
	api.HandleFunc("/consultores", consultorHandler.ListarConsultores).Methods("GET")
	api.HandleFunc("/consultores/{id}", consultorHandler.BuscarPorID).Methods("GET")
	api.HandleFunc("/consultores/{id}", consultorHandler.AtualizarConsultor).Methods("PUT")
	api.HandleFunc("/consultores/{id}", consultorHandler.DeletarConsultor).Methods("DELETE")
	api.HandleFunc("/consultores/{id}/resumo", consultorHandler.ObterResumoConsultor).Methods("GET")
	api.HandleFunc("/consultores/me", consultorHandler.Me).Methods("GET")

	// Negociações
	negHandler := negociacao.NewHandler(db)
	api.HandleFunc("/negociacoes", negHandler.Criar).Methods("POST")
	api.HandleFunc("/negociacoes/{id}", negHandler.BuscarPorID).Methods("GET")
	api.HandleFunc("/consultores/{id}/negociacoes", negHandler.ListarPorConsultor).Methods("GET")
	api.HandleFunc("/negociacoes/{id}", negHandler.Atualizar).Methods("PUT")
	api.HandleFunc("/negociacoes/{id}", negHandler.Deletar).Methods("DELETE")

	// Contratos
	contratoHandler := contrato.NewHandler(db)
	api.HandleFunc("/negociacoes/{id}/contrato", contratoHandler.CriarParaNegociacao).Methods("POST")
	api.HandleFunc("/negociacoes/{id}/contrato", contratoHandler.BuscarPorNegociacao).Methods("GET")
	api.HandleFunc("/consultores/{id}/contratos", contratoHandler.ListarPorConsultor).Methods("GET")
	api.HandleFunc("/contratos/{id}", contratoHandler.Atualizar).Methods("PUT")
	api.HandleFunc("/contratos/{id}", contratoHandler.Deletar).Methods("DELETE")

	// Comentários
	comentHandler := comentario.NewHandler(db)
	api.HandleFunc("/negociacoes/{id}/comentarios", comentHandler.ListarPorNegociacao).Methods("GET")
	api.HandleFunc("/negociacoes/{id}/comentarios", comentHandler.CriarComentario).Methods("POST")
	api.HandleFunc("/comentarios/{id}", comentHandler.BuscarPorID).Methods("GET")
	api.HandleFunc("/comentarios/{id}", comentHandler.Atualizar).Methods("PUT")
	api.HandleFunc("/comentarios/{id}", comentHandler.RemoverComentario).Methods("DELETE")

	fmt.Println("Servidor rodando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
