package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/KromaEnergia/api-consultor/internal/auth"
	"github.com/KromaEnergia/api-consultor/internal/comentario"
	"github.com/KromaEnergia/api-consultor/internal/consultor"
	"github.com/KromaEnergia/api-consultor/internal/contrato"
	"github.com/KromaEnergia/api-consultor/internal/negociacao"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Conexão ao banco com retry para aguardar readiness
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "host=db user=postgres password=postgres dbname=consultor port=5432 sslmode=disable"
	}

	var db *gorm.DB
	var err error
	for i := 1; i <= 5; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		// Se for erro de startup do Postgres, aguarda e tenta novamente
		if strings.Contains(err.Error(), "starting up") {
			log.Printf("Banco ainda iniciando (tentativa %d/5), aguardando 2s...\n", i)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
	if err != nil {
		log.Fatal("Erro ao conectar no banco: ", err)
	}

	// AutoMigrate modelos
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

	// Rotas públicas
	r.HandleFunc("/login", consultorHandler.Login).Methods("POST")
	r.HandleFunc("/consultores", consultorHandler.CriarConsultor).Methods("POST")

	// Subrouter protegido por JWT
	authRoutes := r.NewRoute().Subrouter()
	authRoutes.Use(auth.MiddlewareAutenticacao)

	// Consultores
	authRoutes.HandleFunc("/consultores", consultorHandler.ListarConsultores).Methods("GET")
	authRoutes.HandleFunc("/consultores/me", consultorHandler.Me).Methods("GET")
	authRoutes.HandleFunc("/consultores/{id:[0-9]+}", consultorHandler.BuscarPorID).Methods("GET")
	authRoutes.HandleFunc("/consultores/{id:[0-9]+}", consultorHandler.AtualizarConsultor).Methods("PUT")
	authRoutes.HandleFunc("/consultores/{id:[0-9]+}", consultorHandler.DeletarConsultor).Methods("DELETE")
	authRoutes.HandleFunc("/consultores/{id:[0-9]+}/resumo", consultorHandler.ObterResumoConsultor).Methods("GET")
	authRoutes.HandleFunc("/consultores/{id:[0-9]+}/solicitar-cnpj", consultorHandler.SolicitarAlteracaoCNPJ).Methods("PUT")
	authRoutes.HandleFunc("/consultores/{id:[0-9]+}/gerenciar-cnpj", consultorHandler.GerenciarAlteracaoCNPJ).Methods("POST")
	authRoutes.HandleFunc("/consultores/{id:[0-9]+}/termo-parceria", consultorHandler.AtualizarTermoDeParceria).Methods("PUT")
	authRoutes.HandleFunc("/consultores/me", consultorHandler.AtualizarMeuPerfil).Methods("PUT")
	authRoutes.HandleFunc("/consultores/{id:[0-9]+}/solicitar-email", consultorHandler.SolicitarAlteracaoEmail).Methods("PUT")
	authRoutes.HandleFunc("/consultores/{id:[0-9]+}/gerenciar-email", consultorHandler.GerenciarAlteracaoEmail).Methods("POST")

	// 2) Rotas dinâmicas com ID (usadas por admins ou para outros fins)
	authRoutes.HandleFunc("/consultores/{id:[0-9]+}", consultorHandler.BuscarPorID).Methods("GET")
	authRoutes.HandleFunc("/consultores/{id:[0-9]+}", consultorHandler.AtualizarConsultor).Methods("PUT")

	// Negociações
	negHandler := negociacao.NewHandler(db)
	authRoutes.HandleFunc("/negociacoes", negHandler.Criar).Methods("POST")
	authRoutes.HandleFunc("/negociacoes/{id}", negHandler.BuscarPorID).Methods("GET")
	authRoutes.HandleFunc("/consultores/{id}/negociacoes", negHandler.ListarPorConsultor).Methods("GET")
	authRoutes.HandleFunc("/negociacoes/{id}", negHandler.Atualizar).Methods("PUT")
	authRoutes.HandleFunc("/negociacoes/{id}", negHandler.Deletar).Methods("DELETE")
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/arquivos", negHandler.AdicionarArquivos).Methods("POST")

	// Contratos
	contratoHandler := contrato.NewHandler(db)
	authRoutes.HandleFunc("/negociacoes/{id}/contrato", contratoHandler.CriarParaNegociacao).Methods("POST")
	authRoutes.HandleFunc("/negociacoes/{id}/contrato", contratoHandler.BuscarPorNegociacao).Methods("GET")
	authRoutes.HandleFunc("/consultores/{id}/contratos", contratoHandler.ListarPorConsultor).Methods("GET")
	authRoutes.HandleFunc("/contratos/{id}", contratoHandler.Atualizar).Methods("PUT")
	authRoutes.HandleFunc("/contratos/{id}", contratoHandler.Deletar).Methods("DELETE")

	// **Rota de COMISSÕES** — só pode vir *depois* do contratoHandler ser declarado**
	authRoutes.HandleFunc("/comissoes", contratoHandler.Comissoes).Methods("GET")

	// Comentários
	comentHandler := comentario.NewHandler(db)
	authRoutes.HandleFunc("/negociacoes/{id}/comentarios", comentHandler.ListarPorNegociacao).Methods("GET")
	authRoutes.HandleFunc("/negociacoes/{id}/comentarios", comentHandler.CriarComentario).Methods("POST")
	authRoutes.HandleFunc("/comentarios/{id}", comentHandler.BuscarPorID).Methods("GET")
	authRoutes.HandleFunc("/comentarios/{id}", comentHandler.Atualizar).Methods("PUT")
	authRoutes.HandleFunc("/comentarios/{id}", comentHandler.RemoverComentario).Methods("DELETE")

	// CORS global: liberar todas origens, métodos e headers
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Authorization"},
		AllowCredentials: false,
	})

	handler := c.Handler(r)
	fmt.Println("Servidor rodando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
