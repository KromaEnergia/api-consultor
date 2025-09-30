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

	// -------- Drop & Migrate (apenas se você realmente quer limpar em startup) --------
	// log.Println("[AVISO] APAGANDO TODAS AS TABELAS DO BANCO DE DADOS (startup drop)...")
	// if err := database.Migrator().DropTable(
	// 	&parcelacomissao.ParcelaComissao{},
	// 	&calculocomissao.CalculoComissao{},
	// 	&produtos.Produto{},
	// 	&contrato.Contrato{},
	// 	&models.Comentario{},
	// 	&models.Negociacao{},
	// 	&consultor.Consultor{},
	// 	&comercial.Comercial{},
	// 	&auth.RefreshToken{},
	// ); err != nil {
	// 	log.Fatal("Erro ao apagar tabelas: ", err)
	// }
	// log.Println("[SUCESSO] Todas as tabelas foram apagadas.")

	if err := database.AutoMigrate(
		&comercial.Comercial{},
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

	// Healthcheck
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

	// Admin
	adminRoutes := authRoutes.PathPrefix("").Subrouter()
	adminRoutes.Use(auth.RequireAdmin)
	adminRoutes.HandleFunc("/comerciais", comercialHandler.List).Methods("GET")
	adminRoutes.HandleFunc("/comerciais/{id:[0-9]+}", comercialHandler.Update).Methods("PUT")
	adminRoutes.HandleFunc("/comerciais/{id:[0-9]+}", comercialHandler.Delete).Methods("DELETE")

	// Comercial (autenticado)
	authRoutes.HandleFunc("/comerciais/me", comercialHandler.Me).Methods("GET")
	authRoutes.HandleFunc("/comerciais/{id:[0-9]+}", comercialHandler.GetByID).Methods("GET")

	// Consultores (autenticado)
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
	consultorRoutes.HandleFunc("/{id:[0-9]+}/dados-bancarios", consultorHandler.GetDadosBancariosHandler).Methods("GET")
	consultorRoutes.HandleFunc("/{id:[0-9]+}/dados-bancarios", consultorHandler.UpdateDadosBancariosHandler).Methods("PUT")
	consultorRoutes.HandleFunc("/{id:[0-9]+}/dados-bancarios", consultorHandler.DeleteDadosBancariosHandler).Methods("DELETE")

	// -------- Negociações --------
	authRoutes.HandleFunc("/negociacoes", negHandler.Criar).Methods("POST")
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}", negHandler.BuscarPorID).Methods("GET")
	authRoutes.HandleFunc("/consultores/{id:[0-9]+}/negociacoes", negHandler.ListarPorConsultor).Methods("GET")
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}", negHandler.Atualizar).Methods("PUT")
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}", negHandler.Deletar).Methods("DELETE")

	// Arquivos livres (array genérico)
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/arquivos", negHandler.AdicionarArquivos).Methods("POST") // body: { "urls": ["...","..."] }
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/arquivos/{idx:[0-9]+}", negHandler.RemoverArquivo).Methods("DELETE")

	// Status da negociação
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/status", negHandler.AtualizarStatus).Methods("PATCH") // body: { "status": "..." }

	// ===== Anexos SIMPLES (url + status opcional) =====
	// Body esperado: { "url": "https://...", "status": true }  // "status" é opcional; se omitido, só atualiza a URL
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/logo", negHandler.PatchLogo).Methods("PATCH")
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/anexo-estudo", negHandler.PatchAnexoEstudo).Methods("PATCH")
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/contrato-kc", negHandler.PatchContratoKC).Methods("PATCH")
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/anexo-contrato-social", negHandler.PatchAnexoContratoSocial).Methods("PATCH")
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/anexo-procuracao", negHandler.PatchAnexoProcuracao).Methods("PATCH")
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/anexo-representante-legal", negHandler.PatchAnexoRepresentanteLegal).Methods("PATCH")
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/anexo-estudo-viabilidade", negHandler.PatchAnexoEstudoDeViabilidade).Methods("PATCH")

	// ===== Anexo Fatura (múltiplos itens + status textual) =====
	// Adicionar item: body: { "url": "https://..." }
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/anexo-fatura/itens", negHandler.PostFaturaItem).Methods("POST")
	// Remover item por índice (0-based)
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/anexo-fatura/itens/{idx:[0-9]+}", negHandler.DeleteFaturaItem).Methods("DELETE")
	// Atualizar status textual: body: { "status": "Pendente" | "Enviado" | "Aprovado" | ... }
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/anexo-fatura/status", negHandler.PatchFaturaStatus).Methods("PATCH")

	// -------- Produtos --------
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/produtos", prodHandler.CreateProdutos).Methods("POST")
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/produtos", prodHandler.ListProdutos).Methods("GET")
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/produtos/{pid:[0-9]+}", prodHandler.GetProduto).Methods("GET")
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/produtos/{pid:[0-9]+}", prodHandler.UpdateProduto).Methods("PUT")
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/produtos/{pid:[0-9]+}", prodHandler.DeleteProduto).Methods("DELETE")

	// -------- Contratos --------
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/contrato", contratoHandler.CriarParaNegociacao).Methods("POST")
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/contrato", contratoHandler.BuscarPorNegociacao).Methods("GET")
	authRoutes.HandleFunc("/consultores/{id:[0-9]+}/contratos", contratoHandler.ListarPorConsultor).Methods("GET")
	authRoutes.HandleFunc("/contratos/{id:[0-9]+}", contratoHandler.Atualizar).Methods("PUT")
	authRoutes.HandleFunc("/contratos/{id:[0-9]+}", contratoHandler.Deletar).Methods("DELETE")

	// -------- Comentários --------
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/comentarios", comentHandler.ListarPorNegociacao).Methods("GET")
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/comentarios", comentHandler.CriarComentario).Methods("POST")
	authRoutes.HandleFunc("/comentarios/{id:[0-9]+}", comentHandler.BuscarPorID).Methods("GET")
	authRoutes.HandleFunc("/comentarios/{id:[0-9]+}", comentHandler.Atualizar).Methods("PUT")
	authRoutes.HandleFunc("/comentarios/{id:[0-9]+}", comentHandler.RemoverComentario).Methods("DELETE")

	// -------- Cálculo de comissão --------
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/calculos-comissao", calcHandler.Create).Methods("POST")
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/calculos-comissao", calcHandler.List).Methods("GET")
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/calculos-comissao/{cid:[0-9]+}", calcHandler.Get).Methods("GET")
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/calculos-comissao/{cid:[0-9]+}", calcHandler.Update).Methods("PUT")
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/calculos-comissao/{cid:[0-9]+}", calcHandler.Delete).Methods("DELETE")
	authRoutes.HandleFunc("/negociacoes/{id:[0-9]+}/calculos-comissao/{cid:[0-9]+}/status", calcHandler.UpdateStatus).Methods("PATCH")

	// -------- Parcelas de comissão --------
	// criar/listar parcelas de um cálculo
	authRoutes.HandleFunc("/calculos-comissao/{cid:[0-9]+}/parcelas", parcelaHandler.CreateForCalculo).Methods(http.MethodPost)
	authRoutes.HandleFunc("/calculos-comissao/{cid:[0-9]+}/parcelas", parcelaHandler.List).Methods(http.MethodGet)
	// operações sobre uma parcela específica
	authRoutes.HandleFunc("/parcelas/{pid:[0-9]+}", parcelaHandler.Update).Methods(http.MethodPut)
	authRoutes.HandleFunc("/parcelas/{pid:[0-9]+}", parcelaHandler.Delete).Methods(http.MethodDelete)
	authRoutes.HandleFunc("/parcelas/{pid:[0-9]+}/status", parcelaHandler.UpdateStatus).Methods(http.MethodPatch)
	// anexo/nota fiscal da parcela
	authRoutes.HandleFunc("/parcelas/{pid:[0-9]+}/anexo", parcelaHandler.UpdateAnexo).Methods(http.MethodPost) // body: { "url": "..." }
	authRoutes.HandleFunc("/parcelas/{pid:[0-9]+}/anexo", parcelaHandler.DeleteAnexo).Methods(http.MethodDelete)
	authRoutes.HandleFunc("/parcelas/{pid:[0-9]+}/nota-fiscal", parcelaHandler.UpdateNotaFiscal).Methods(http.MethodPost) // body: { "url": "..." }
	authRoutes.HandleFunc("/parcelas/{pid:[0-9]+}/nota-fiscal", parcelaHandler.DeleteNotaFiscal).Methods(http.MethodDelete)

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
