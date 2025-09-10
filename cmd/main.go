package main

import (
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
	"github.com/KromaEnergia/api-consultor/internal/utils/db"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"gorm.io/gorm"
)

func main() {
	// ====== MODO SEM BANCO (para teste no App Runner) ======
	if os.Getenv("SKIP_DB") == "1" {
		r := mux.NewRouter()
		r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}).Methods(http.MethodGet)

		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		log.Printf("[WARN] SKIP_DB=1: subindo só /health em :%s", port)
		log.Fatal(http.ListenAndServe(":"+port, r))
		return
	}
	// ====== FIM DO MODO SEM BANCO ======

	// -------- Conexão ao banco com retry --------
	var database *gorm.DB
	var err error
	for i := 1; i <= 10; i++ {
		database, err = db.GetDB()
		if err == nil {
			break
		}
		if strings.Contains(err.Error(), "starting up") || strings.Contains(err.Error(), "connection refused") {
			log.Printf("Banco indisponível (tentativa %d/10): %v. Aguardando 2s...", i, err)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
	if err != nil {
		log.Fatal("Erro ao conectar no banco: ", err)
	}

	// -------- DropTable só em DEV --------
	if os.Getenv("DEV_DROP_ALL") == "1" {
		log.Println("[DEV] APAGANDO TODAS AS TABELAS... (DEV_DROP_ALL=1)")
		if err := database.Migrator().DropTable(
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
	}

	// -------- AutoMigrate --------
	if err := database.AutoMigrate(
		&comercial.Comercial{},
	); err != nil {
		log.Fatal("Erro no AutoMigrate: comercial ", err)
	}

	// -------- AutoMigrate --------
	if err := database.AutoMigrate(

		&consultor.Consultor{},
		&models.Negociacao{},
		&models.Comentario{},
		&contrato.Contrato{},
		&produtos.Produto{},
		&calculocomissao.CalculoComissao{},
		&parcelacomissao.ParcelaComissao{},
		&auth.RefreshToken{},
	); err != nil {
		log.Fatal("Erro no AutoMigrate: ", err)
	}

	// -------- Instancia handlers/repos --------
	consultorHandler := consultor.NewHandler(database)
	comercialHandler := comercial.NewHandler(database)
	negHandler := negociacao.NewHandler(database)
	contratoHandler := contrato.NewHandler(database)

	prodRepo := produtos.NewRepository(database)
	prodHandler := produtos.NewHandler(prodRepo)

	calcRepo := calculocomissao.NewRepository(database)
	calcHandler := calculocomissao.NewHandler(calcRepo)

	parcelaRepo := parcelacomissao.NewRepository(database)
	parcelaHandler := parcelacomissao.NewHandler(parcelaRepo)

	comentHandler := comentario.NewHandler(database)

	// -------- Router --------
	r := mux.NewRouter()

	// Healthcheck (para App Runner/ECS)
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}).Methods(http.MethodGet)

	// ---------- Rotas públicas ----------
	r.HandleFunc("/.well-known/jwks.json", auth.JWKSHandler).Methods("GET")
	r.HandleFunc("/auth/refresh", auth.RefreshHTTPHandler(database)).Methods("POST")
	r.HandleFunc("/auth/logout", auth.LogoutHTTPHandler(database)).Methods("POST")
	r.HandleFunc("/consultores/login", consultorHandler.Login).Methods("POST")
	r.HandleFunc("/consultores", consultorHandler.CriarConsultor).Methods("POST")
	r.HandleFunc("/comerciais/login", comercialHandler.Login).Methods("POST")
	r.HandleFunc("/comerciais", comercialHandler.Create).Methods("POST")

	// ---------- Rotas protegidas ----------
	authRoutes := r.PathPrefix("").Subrouter()
	authRoutes.Use(auth.MiddlewareAutenticacao)

	adminRoutes := authRoutes.PathPrefix("").Subrouter()
	adminRoutes.Use(auth.RequireAdmin)
	adminRoutes.HandleFunc("/comerciais", comercialHandler.List).Methods("GET")
	adminRoutes.HandleFunc("/comerciais/{id:[0-9]+}", comercialHandler.Update).Methods("PUT")
	adminRoutes.HandleFunc("/comerciais/{id:[0-9]+}", comercialHandler.Delete).Methods("DELETE")

	authRoutes.HandleFunc("/comerciais/me", comercialHandler.Me).Methods("GET")
	authRoutes.HandleFunc("/comerciais/{id:[0-9]+}", comercialHandler.GetByID).Methods("GET")

	// Consultores
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
	// criar/listar parcelas de um cálculo
	authRoutes.HandleFunc("/calculos-comissao/{cid:[0-9]+}/parcelas", parcelaHandler.CreateForCalculo).Methods(http.MethodPost)
	authRoutes.HandleFunc("/calculos-comissao/{cid:[0-9]+}/parcelas", parcelaHandler.List).Methods(http.MethodGet)

	// operações sobre uma parcela específica
	authRoutes.HandleFunc("/parcelas/{pid:[0-9]+}", parcelaHandler.Update).Methods(http.MethodPut)
	authRoutes.HandleFunc("/parcelas/{pid:[0-9]+}", parcelaHandler.Delete).Methods(http.MethodDelete)
	authRoutes.HandleFunc("/parcelas/{pid:[0-9]+}/status", parcelaHandler.UpdateStatus).Methods(http.MethodPatch)
	authRoutes.HandleFunc("/parcelas/{pid:[0-9]+}/anexo", parcelaHandler.UpdateAnexo).Methods(http.MethodPost)
	authRoutes.HandleFunc("/parcelas/{pid:[0-9]+}/anexo", parcelaHandler.DeleteAnexo).Methods(http.MethodDelete)

	// -------- CORS --------
	origins := os.Getenv("ALLOWED_ORIGINS")
	if origins == "" {
		origins = "http://localhost:3000"
	}
	allowed := strings.Split(origins, ",")
	c := cors.New(cors.Options{
		AllowedOrigins:   allowed,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Authorization"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Servidor rodando em :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
