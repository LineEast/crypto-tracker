package server

import (
	"github.com/fasthttp/router"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valyala/fasthttp"

	aheaders "github.com/go-asphyxia/http/headers"
	amethods "github.com/go-asphyxia/http/methods"
	"github.com/go-asphyxia/middlewares/CORS"
)

type (
	Server struct {
		router *router.Router
		cors   *CORS.CORS
	}
)

func New(db *pgxpool.Pool) *Server {
	server := &Server{
		cors: CORS.NewCORS(&CORS.Configuration{
			Origins: nil,
			Methods: []string{amethods.GET, amethods.POST, amethods.OPTIONS}, //, amethods.PUT, amethods.DELETE}
			Headers: []string{aheaders.Accept, aheaders.ContentType},
		}),
	}

	r := router.New()

	r.OPTIONS("/{everything:*}", server.cors.Handler())

	api := r.Group("/api")

	btcusdt := api.Group("/btcusdt")
	btcusdt.GET("", btusdtGET(db))
	btcusdt.POST("", btusdtPOST(db))

	currencies := api.Group("/currencies")
	currencies.GET("", currenciesGET(db))
	currencies.POST("", currenciesPOST(db))

	server.router = r

	return server
}

func (s *Server) Run() error {
	return fasthttp.ListenAndServe(":8080", s.cors.Middleware(s.router.Handler))
}
