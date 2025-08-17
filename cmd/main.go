package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/KromaEnergia/api-consultor/internal/auth"
	"github.com/KromaEnergia/api-consultor/internal/calculocomissao"
	"github.com/KromaEnergia/api-consultor/internal/comentario"
	"github.com/KromaEnergia/api-consultor/internal/comercial"
	"github.com/KromaEnergia/api-consultor/internal/consultor"
	"github.com/KromaEnergia/api-consultor/internal/contrato"
	"github.com/KromaEnergia/api-consultor/internal/models"
	"github.com/KromaEnergia/api-consultor/internal/negociacao"
	"github.com/KromaEnergia/api-consultor/internal/parcelacomissao"
	"github.com/KromaEnergia/api-consultor/internal/produtos"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Conexão ao banco com retry
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

	// *** CUIDADO: ISSO APAGA TODAS AS TABELAS ***
	// Deixe apenas em ambiente de DEV quando precisar resetar.

	log.Println("[AVISO] APAGANDO TODAS AS TABELAS DO BANCO DE DADOS...")
	if err := db.Migrator().DropTable(
		&parcelacomissao.ParcelaComissao{},
		&calculocomissao.CalculoComissao{},
		&produtos.Produto{},
		&contrato.Contrato{},
		&models.Comentario{},
		&models.Negociacao{},
		&consultor.Consultor{},
		&comercial.Comercial{},
		&auth.RefreshToken{},
	); err != nil {
		log.Fatal("Erro ao apagar tabelas: ", err)
	}
	log.Println("[SUCESSO] Todas as tabelas foram apagadas.")

	// AutoMigrate modelos (INCLUI o RefreshToken)
	if err := db.AutoMigrate(
		&consultor.Consultor{},
		&comercial.Comercial{},
		&models.Negociacao{},
		&models.Comentario{},
		&contrato.Contrato{},
		&produtos.Produto{},
		&calculocomissao.CalculoComissao{},
		&parcelacomissao.ParcelaComissao{},
		&auth.RefreshToken{}, // <<< AQUI
	); err != nil {
		log.Fatal("Erro no AutoMigrate: ", err)
	}

	// Instancia handlers/repos
	consultorHandler := consultor.NewHandler(db)
	comercialHandler := comercial.NewHandler(db)
	negHandler := negociacao.NewHandler(db)
	contratoHandler := contrato.NewHandler(db)

	prodRepo := produtos.NewRepository(db)
	prodHandler := produtos.NewHandler(prodRepo)

	calcRepo := calculocomissao.NewRepository(db)
	calcHandler := calculocomissao.NewHandler(calcRepo)

	parcelaRepo := parcelacomissao.NewRepository(db)
	parcelaHandler := parcelacomissao.NewHandler(parcelaRepo)
	parcelasHandler := parcelacomissao.NewHandler(parcelaRepo)

	comissoesHandler := consultor.NewComissoesHandler(db)
	comentHandler := comentario.NewHandler(db)

	// Router
	r := mux.NewRouter()

	// ---------- Rotas públicas ----------
	// JWKS (chaves públicas do JWT RS256)
	r.HandleFunc("/.well-known/jwks.json", auth.JWKSHandler).Methods("GET")

	// Auth (refresh/logout)
	r.HandleFunc("/auth/refresh", auth.RefreshHTTPHandler(db)).Methods("POST")
	r.HandleFunc("/auth/logout", auth.LogoutHTTPHandler(db)).Methods("POST")

	// Logins e criação (públicas)
	r.HandleFunc("/consultores/login", consultorHandler.Login).Methods("POST")
	r.HandleFunc("/consultores", consultorHandler.CriarConsultor).Methods("POST")

	r.HandleFunc("/comerciais/login", comercialHandler.Login).Methods("POST")
	r.HandleFunc("/comerciais", comercialHandler.Create).Methods("POST")

	// ---------- Rotas protegidas ----------
	authRoutes := r.PathPrefix("").Subrouter()
	authRoutes.Use(auth.MiddlewareAutenticacao)

	// Admin-only (Comercial manda em tudo)
	adminRoutes := authRoutes.PathPrefix("").Subrouter()
	adminRoutes.Use(auth.RequireAdmin)
	adminRoutes.HandleFunc("/comerciais", comercialHandler.List).Methods("GET")
	adminRoutes.HandleFunc("/comerciais/{id:[0-9]+}", comercialHandler.Update).Methods("PUT")
	adminRoutes.HandleFunc("/comerciais/{id:[0-9]+}", comercialHandler.Delete).Methods("DELETE")

	// Comercial - rotas não admin (ex.: me)
	authRoutes.HandleFunc("/comerciais/me", comercialHandler.Me).Methods("GET")
	authRoutes.HandleFunc("/comerciais/{id:[0-9]+}", comercialHandler.GetByID).Methods("GET")

	// Consultores (subrouter com auth)
	consultorRoutes := authRoutes.PathPrefix("/consultores").Subrouter()
	consultorRoutes.HandleFunc("/me", consultorHandler.Me).Methods("GET")
	consultorRoutes.HandleFunc("", consultorHandler.ListarConsultores).Methods("GET")
	consultorRoutes.HandleFunc("/{id:[0-9]+}", consultorHandler.BuscarPorID).Methods("GET")
	consultorRoutes.HandleFunc("/me", consultorHandler.AtualizarMeuPerfil).Methods("PUT")
	consultorRoutes.HandleFunc("/{id:[0-9]+}", consultorHandler.AtualizarConsultor).Methods("PUT")
	consultorRoutes.HandleFunc("/{id:[0-9]+}", consultorHandler.DeletarConsultor).Methods("DELETE")
	consultorRoutes.HandleFunc("/{id:[0-9]+}/resumo", consultorHandler.ObterResumoConsultor).Methods("GET")
	consultorRoutes.HandleFunc("/{id:[0-9]+}/solicitar-cnpj", consultorHandler.SolicitarAlteracaoCNPJ).Methods("PUT")
	consultorRoutes.HandleFunc("/{id:[0-9]+}/gerenciar-cnpj", consultorHandler.GerenciarAlteracaoCNPJ).Methods("POST")
	consultorRoutes.HandleFunc("/{id:[0-9]+}/termo-parceria", consultorHandler.AtualizarTermoDeParceria).Methods("PUT")
	consultorRoutes.HandleFunc("/{id:[0-9]+}/solicitar-email", consultorHandler.SolicitarAlteracaoEmail).Methods("PUT")
	consultorRoutes.HandleFunc("/{id:[0-9]+}/gerenciar-email", consultorHandler.GerenciarAlteracaoEmail).Methods("POST")
	consultorRoutes.HandleFunc("/", consultorHandler.ListarConsultoresSimples).Methods("GET")
	consultorRoutes.HandleFunc("/completo", consultorHandler.ListarConsultoresCompletos).Methods("GET")
	consultorRoutes.HandleFunc("/{id}/dados-bancarios", consultorHandler.GetDadosBancariosHandler).Methods("GET")
	consultorRoutes.HandleFunc("/{id}/dados-bancarios", consultorHandler.UpdateDadosBancariosHandler).Methods("PUT")
	consultorRoutes.HandleFunc("/{id}/dados-bancarios", consultorHandler.DeleteDadosBancariosHandler).Methods("DELETE")
	consultorRoutes.HandleFunc("/comissoes", comissoesHandler.GetResumo).Methods("GET")

	// Negociações
	authRoutes.HandleFunc("/negociacoes", negHandler.Criar).Methods("POST")
	authRoutes.HandleFunc("/negociacoes/{id}", negHandler.BuscarPorID).Methods("GET")
	authRoutes.HandleFunc("/consultores/{id}/negociacoes", negHandler.ListarPorConsultor).Methods("GET")
	authRoutes.HandleFunc("/negociacoes/{id}", negHandler.Atualizar).Methods("PUT")
	authRoutes.HandleFunc("/negociacoes/{id}", negHandler.Deletar).Methods("DELETE")
	authRoutes.HandleFunc("/negociacoes/{id}/arquivos", negHandler.AdicionarArquivos).Methods("POST")
	authRoutes.HandleFunc("/negociacoes/{id}/arquivos/{idx}", negHandler.RemoverArquivo).Methods("DELETE")
	authRoutes.HandleFunc("/negociacoes/{id}/status", negHandler.AtualizarStatus).Methods("PATCH")
	authRoutes.HandleFunc("/negociacoes/{id}/anexo-estudo", negHandler.PatchAnexoEstudo).Methods("PATCH")
	authRoutes.HandleFunc("/negociacoes/{id}/contrato-kc", negHandler.PatchContratoKC).Methods("PATCH")

	// Produtos
	authRoutes.HandleFunc("/negociacoes/{id}/produtos", prodHandler.CreateProdutos).Methods("POST")
	authRoutes.HandleFunc("/negociacoes/{id}/produtos", prodHandler.ListProdutos).Methods("GET")
	authRoutes.HandleFunc("/negociacoes/{id}/produtos/{pid}", prodHandler.GetProduto).Methods("GET")
	authRoutes.HandleFunc("/negociacoes/{id}/produtos/{pid}", prodHandler.UpdateProduto).Methods("PUT")
	authRoutes.HandleFunc("/negociacoes/{id}/produtos/{pid}", prodHandler.DeleteProduto).Methods("DELETE")

	// Contratos
	authRoutes.HandleFunc("/negociacoes/{id}/contrato", contratoHandler.CriarParaNegociacao).Methods("POST")
	authRoutes.HandleFunc("/negociacoes/{id}/contrato", contratoHandler.BuscarPorNegociacao).Methods("GET")
	authRoutes.HandleFunc("/consultores/{id}/contratos", contratoHandler.ListarPorConsultor).Methods("GET")
	authRoutes.HandleFunc("/contratos/{id}", contratoHandler.Atualizar).Methods("PUT")
	authRoutes.HandleFunc("/contratos/{id}", contratoHandler.Deletar).Methods("DELETE")

	// Comentários
	authRoutes.HandleFunc("/negociacoes/{id}/comentarios", comentHandler.ListarPorNegociacao).Methods("GET")
	authRoutes.HandleFunc("/negociacoes/{id}/comentarios", comentHandler.CriarComentario).Methods("POST")
	authRoutes.HandleFunc("/comentarios/{id}", comentHandler.BuscarPorID).Methods("GET")
	authRoutes.HandleFunc("/comentarios/{id}", comentHandler.Atualizar).Methods("PUT")
	authRoutes.HandleFunc("/comentarios/{id}", comentHandler.RemoverComentario).Methods("DELETE")

	// Cálculo de comissão
	authRoutes.HandleFunc("/negociacoes/{id}/calculos-comissao", calcHandler.Create).Methods("POST")
	authRoutes.HandleFunc("/negociacoes/{id}/calculos-comissao", calcHandler.List).Methods("GET")
	authRoutes.HandleFunc("/negociacoes/{id}/calculos-comissao/{cid}", calcHandler.Get).Methods("GET")
	authRoutes.HandleFunc("/negociacoes/{id}/calculos-comissao/{cid}", calcHandler.Update).Methods("PUT")
	authRoutes.HandleFunc("/negociacoes/{id}/calculos-comissao/{cid}", calcHandler.Delete).Methods("DELETE")
	authRoutes.HandleFunc("/negociacoes/{id}/calculos-comissao/{cid}/status", calcHandler.UpdateStatus).Methods("PATCH")

	// Parcelas de comissão
	authRoutes.HandleFunc("/calculos-comissao/{cid}/parcelas", parcelaHandler.List).Methods("GET")
	authRoutes.HandleFunc("/parcelas/{pid}/status", parcelaHandler.UpdateStatus).Methods("PATCH")
	authRoutes.HandleFunc("/parcelas/{pid}/anexo", parcelaHandler.UpdateAnexo).Methods("POST")
	authRoutes.HandleFunc("/parcelas/{pid}/anexo", parcelaHandler.DeleteAnexo).Methods("DELETE")
	authRoutes.HandleFunc("/calculos-comissao/{cid}/parcelas", parcelasHandler.List).Methods("GET")
	authRoutes.HandleFunc("/parcelas/{pid}/status", parcelasHandler.UpdateStatus).Methods("PATCH")
	authRoutes.HandleFunc("/parcelas/{pid}", parcelaHandler.Update).Methods("PUT")

	// CORS (ajuste AllowedOrigins p/ seu front se usar cookies httpOnly no refresh)
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Authorization"},
		AllowCredentials: true, // p/ enviar cookie do refresh
	})

	handler := c.Handler(r)

	fmt.Println("Servidor rodando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
