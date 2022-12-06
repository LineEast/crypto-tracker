package collector

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valyala/fasthttp"
)

type (
	Collector struct {
		DB     *pgxpool.Pool
		config *Config

		FiatClient   *fasthttp.HostClient
		CryptoClient *fasthttp.HostClient
	}

	Config struct {
		FiatEndPoint   string
		CryptoEndPoint string
	}
)

func New(db *pgxpool.Pool, config *Config) *Collector {
	return &Collector{
		DB:     db,
		config: config,

		FiatClient: &fasthttp.HostClient{
			Addr:                  config.FiatEndPoint,
			Name:                  "crypto-tracker-Fiat",
			DialDualStack:         true,
			IsTLS:                 true,
			SecureErrorLogMessage: true,
		},

		CryptoClient: &fasthttp.HostClient{
			Addr:                  config.CryptoEndPoint,
			Name:                  "crypto-tracker-Crypto",
			DialDualStack:         true,
			IsTLS:                 true,
			SecureErrorLogMessage: true,
		},
	}
}

func (collector *Collector) Run() (err error) {

	return
}
