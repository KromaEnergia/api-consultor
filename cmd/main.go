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

	// Para APAGAR o banco, comente o AutoMigrate e descomente este bloco:
	log.Println("[AVISO] APAGANDO TODAS AS TABELAS DO BANCO DE DADOS...")
	if err := db.Migrator().DropTable(
		// A ordem aqui é importante para chaves estrangeiras.
		// GORM tenta lidar com isso, mas é mais seguro apagar as tabelas
		// que têm dependências primeiro (as "filhas").
		&parcelacomissao.ParcelaComissao{},
		&calculocomissao.CalculoComissao{},
		&produtos.Produto{},
		&contrato.Contrato{},
		&models.Comentario{},
		&models.Negociacao{},
		&consultor.Consultor{},
		&comercial.Comercial{},
	); err != nil {
		log.Fatal("Erro ao apagar tabelas: ", err)
	}
	log.Println("[SUCESSO] Todas as tabelas foram apagadas.")

	// AutoMigrate modelos
	if err := db.AutoMigrate(
		&consultor.Consultor{},
		&comercial.Comercial{},
		&models.Negociacao{},
		&models.Comentario{},
		&contrato.Contrato{},
		&produtos.Produto{},
		&calculocomissao.CalculoComissao{},
		&parcelacomissao.ParcelaComissao{},
	); err != nil {
		log.Fatal("Erro no AutoMigrate: ", err)
	}

	// Instancia repositórios e handlers
	comissoesHandler := consultor.NewComissoesHandler(db)
	prodRepo := produtos.NewRepository(db)
	prodHandler := produtos.NewHandler(prodRepo)
	calcRepo := calculocomissao.NewRepository(db)
	calcHandler := calculocomissao.NewHandler(calcRepo)
	parcelasRepo := parcelacomissao.NewRepository(db)           // 3a. Repo
	parcelasHandler := parcelacomissao.NewHandler(parcelasRepo) // 3b. Handler
	r := mux.NewRouter()

	// Handlers de Consultor
	consultorHandler := consultor.NewHandler(db)
	r.HandleFunc("/consultores/login", consultorHandler.Login).Methods("POST")
	r.HandleFunc("/consultores", consultorHandler.CriarConsultor).Methods("POST")

	// Handlers de Comercial
	comercialHandler := comercial.NewHandler(db)
	r.HandleFunc("/comerciais/login", comercialHandler.Login).Methods("POST")
	r.HandleFunc("/comerciais", comercialHandler.Create).Methods("POST")

	// Rotas protegidas por JWT
	authRoutes := r.NewRoute().Subrouter()
	authRoutes.Use(auth.MiddlewareAutenticacao)

	// sub‐router apenas para /consultores
	consultorRoutes := r.PathPrefix("/consultores").Subrouter()
	consultorRoutes.Use(auth.MiddlewareAutenticacao)

	parcelaRepo := parcelacomissao.NewRepository(db)
	parcelaHandler := parcelacomissao.NewHandler(parcelaRepo)

	// GET /consultores/me
	consultorRoutes.HandleFunc("/me", consultorHandler.Me).Methods("GET")
	// GET /consultores
	consultorRoutes.HandleFunc("", consultorHandler.ListarConsultores).Methods("GET")
	// GET /consultores/{id}
	consultorRoutes.HandleFunc("/{id:[0-9]+}", consultorHandler.BuscarPorID).Methods("GET")
	// PUT /consultores/me
	consultorRoutes.HandleFunc("/me", consultorHandler.AtualizarMeuPerfil).Methods("PUT")
	// PUT /consultores/{id}
	consultorRoutes.HandleFunc("/{id:[0-9]+}", consultorHandler.AtualizarConsultor).Methods("PUT")
	// DELETE /consultores/{id}
	consultorRoutes.HandleFunc("/{id:[0-9]+}", consultorHandler.DeletarConsultor).Methods("DELETE")
	// GET /consultores/{id}/resumo
	consultorRoutes.HandleFunc("/{id:[0-9]+}/resumo", consultorHandler.ObterResumoConsultor).Methods("GET")
	// PUT /consultores/{id}/solicitar-cnpj
	consultorRoutes.HandleFunc("/{id:[0-9]+}/solicitar-cnpj", consultorHandler.SolicitarAlteracaoCNPJ).Methods("PUT")
	// POST /consultores/{id}/gerenciar-cnpj
	consultorRoutes.HandleFunc("/{id:[0-9]+}/gerenciar-cnpj", consultorHandler.GerenciarAlteracaoCNPJ).Methods("POST")
	// PUT /consultores/{id}/termo-parceria
	consultorRoutes.HandleFunc("/{id:[0-9]+}/termo-parceria", consultorHandler.AtualizarTermoDeParceria).Methods("PUT")
	// PUT /consultores/{id}/solicitar-email
	consultorRoutes.HandleFunc("/{id:[0-9]+}/solicitar-email", consultorHandler.SolicitarAlteracaoEmail).Methods("PUT")
	// POST /consultores/{id}/gerenciar-email
	consultorRoutes.HandleFunc("/{id:[0-9]+}/gerenciar-email", consultorHandler.GerenciarAlteracaoEmail).Methods("POST")
	consultorRoutes.HandleFunc("/", consultorHandler.ListarConsultoresSimples).Methods("GET")
	consultorRoutes.HandleFunc("/completo", consultorHandler.ListarConsultoresCompletos).Methods("GET")

	consultorRoutes.HandleFunc("/{id}/dados-bancarios", consultorHandler.GetDadosBancariosHandler).Methods("GET")
	consultorRoutes.HandleFunc("/{id}/dados-bancarios", consultorHandler.UpdateDadosBancariosHandler).Methods("PUT")
	consultorRoutes.HandleFunc("/{id}/dados-bancarios", consultorHandler.DeleteDadosBancariosHandler).Methods("DELETE")

	// Rotas de Comercial
	authRoutes.HandleFunc("/comerciais", comercialHandler.List).Methods("GET")
	authRoutes.HandleFunc("/comerciais/{id:[0-9]+}", comercialHandler.GetByID).Methods("GET")
	authRoutes.HandleFunc("/comerciais/{id:[0-9]+}", comercialHandler.Update).Methods("PUT")
	authRoutes.HandleFunc("/comerciais/{id:[0-9]+}", comercialHandler.Delete).Methods("DELETE")
	authRoutes.HandleFunc("/comerciais/me", comercialHandler.Me).Methods("GET")

	// Rotas de Negociação
	negHandler := negociacao.NewHandler(db)
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

	// Rotas de Produtos
	authRoutes.HandleFunc("/negociacoes/{id}/produtos", prodHandler.CreateProdutos).Methods("POST")
	authRoutes.HandleFunc("/negociacoes/{id}/produtos", prodHandler.ListProdutos).Methods("GET")
	authRoutes.HandleFunc("/negociacoes/{id}/produtos/{pid}", prodHandler.GetProduto).Methods("GET")
	authRoutes.HandleFunc("/negociacoes/{id}/produtos/{pid}", prodHandler.UpdateProduto).Methods("PUT")
	authRoutes.HandleFunc("/negociacoes/{id}/produtos/{pid}", prodHandler.DeleteProduto).Methods("DELETE")

	// Rotas de Contrato
	contratoHandler := contrato.NewHandler(db)
	authRoutes.HandleFunc("/negociacoes/{id}/contrato", contratoHandler.CriarParaNegociacao).Methods("POST")
	authRoutes.HandleFunc("/negociacoes/{id}/contrato", contratoHandler.BuscarPorNegociacao).Methods("GET")
	authRoutes.HandleFunc("/consultores/{id}/contratos", contratoHandler.ListarPorConsultor).Methods("GET")
	authRoutes.HandleFunc("/contratos/{id}", contratoHandler.Atualizar).Methods("PUT")
	authRoutes.HandleFunc("/contratos/{id}", contratoHandler.Deletar).Methods("DELETE")

	// Rotas de Comentários
	comentHandler := comentario.NewHandler(db)
	authRoutes.HandleFunc("/negociacoes/{id}/comentarios", comentHandler.ListarPorNegociacao).Methods("GET")
	authRoutes.HandleFunc("/negociacoes/{id}/comentarios", comentHandler.CriarComentario).Methods("POST")
	authRoutes.HandleFunc("/comentarios/{id}", comentHandler.BuscarPorID).Methods("GET")
	authRoutes.HandleFunc("/comentarios/{id}", comentHandler.Atualizar).Methods("PUT")
	authRoutes.HandleFunc("/comentarios/{id}", comentHandler.RemoverComentario).Methods("DELETE")

	// --- ROTAS PARA CÁLCULO DE COMISSÃO ---
	authRoutes.HandleFunc("/negociacoes/{id}/calculos-comissao", calcHandler.Create).Methods("POST")
	authRoutes.HandleFunc("/negociacoes/{id}/calculos-comissao", calcHandler.List).Methods("GET")
	authRoutes.HandleFunc("/negociacoes/{id}/calculos-comissao/{cid}", calcHandler.Get).Methods("GET")
	authRoutes.HandleFunc("/negociacoes/{id}/calculos-comissao/{cid}", calcHandler.Update).Methods("PUT")
	authRoutes.HandleFunc("/negociacoes/{id}/calculos-comissao/{cid}", calcHandler.Delete).Methods("DELETE")
	authRoutes.HandleFunc("/negociacoes/{id}/calculos-comissao/{cid}/status", calcHandler.UpdateStatus).Methods("PATCH")

	// --- ROTAS PARA PARCELAS DE COMISSÃO ---
	// Lista todas as parcelas de um cálculo específico
	authRoutes.HandleFunc("/calculos-comissao/{cid}/parcelas", parcelaHandler.List).Methods("GET")

	// Atualiza o status de uma parcela específica
	authRoutes.HandleFunc("/parcelas/{pid}/status", parcelaHandler.UpdateStatus).Methods("PATCH")

	// --- NOVAS ROTAS PARA O ANEXO DA PARCELA ---
	// Adiciona ou atualiza o anexo de uma parcela específica
	authRoutes.HandleFunc("/parcelas/{pid}/anexo", parcelaHandler.UpdateAnexo).Methods("POST")

	// Remove o anexo de uma parcela específica
	authRoutes.HandleFunc("/parcelas/{pid}/anexo", parcelaHandler.DeleteAnexo).Methods("DELETE")

	// <<<---- NOVA ROTA PATCH PARA ATUALIZAR Parcela ---->>>
	authRoutes.HandleFunc("/calculos-comissao/{cid}/parcelas", parcelasHandler.List).Methods("GET")
	// PATCH /parcelas/{pid}/status
	authRoutes.HandleFunc("/parcelas/{pid}/status", parcelasHandler.UpdateStatus).Methods("PATCH")
	authRoutes.HandleFunc("/parcelas/{pid}", parcelaHandler.Update).Methods("PUT")
	consultorRoutes.HandleFunc("/comissoes", comissoesHandler.GetResumo).Methods("GET")

	// CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}, // Adicionado PATCH
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Authorization"},
		AllowCredentials: false,
	})
	handler := c.Handler(r)

	fmt.Println("Servidor rodando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
