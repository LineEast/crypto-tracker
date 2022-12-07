package collector

import (
	"context"
	"encoding/xml"
	"time"

	"github.com/jackc/pgx/v5"
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

	Fiat struct {
		Date   time.Time `xml:"Date"`
		Valute []Valute  `xml:"Valute"`
	}

	Valute struct {
		ID     string  `xml:"-"`
		FiatID string  `xml:"ID"`
		Name   string  `xml:"Name"`
		Code   string  `xml:"CharCode"`
		Value  float64 `xml:"Value"`
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
	go func() {
		for {
			collector.fiat()
		}
	}()

	return
}

func (collector *Collector) fiat() (err error) {
	req := fasthttp.AcquireRequest()
	req.Header.SetMethod("GET")
	req.Header.Set("Content-Type", "application/xml")
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)

	err = collector.FiatClient.Do(req, resp)
	if err != nil {
		return
	}

	fiat := Fiat{}
	err = xml.Unmarshal(resp.Body(), &fiat)
	if err != nil {
		return
	}

	rows, err := collector.DB.Query(
		context.Background(),
		"select id, fiat_id, code from fiat",
	)
	defer rows.Close()

	// в карте те валюты, которые мы уже записывали ранее
	vl := make(map[string]string)
	valute := Valute{}
	for rows.Next() {
		err = rows.Scan(&valute.ID, &valute.FiatID, &valute.Code)
		if err != nil {
			return
		}

		vl[valute.FiatID] = valute.ID
	}

	insertFiat := [][]interface{}{}
	for i := range fiat.Valute {
		if _, ok := vl[fiat.Valute[i].FiatID]; ok {
			insertFiat = append(insertFiat, []interface{}{vl[fiat.Valute[i].FiatID], fiat.Date, fiat.Valute[i].Value})
		} else {
			err = collector.DB.QueryRow(
				context.Background(),
				"insert into fiat (fiat_id, name, code) values ($1, $2, $3) returning id",
				fiat.Valute[i].FiatID, fiat.Valute[i].Name, fiat.Valute[i].Code,
			).Scan(fiat.Valute[i].ID)
			if err != nil {
				return
			}

			insertFiat = append(insertFiat, []interface{}{fiat.Valute[i].ID, fiat.Date, fiat.Valute[i].Value})
		}
	}

	_, err = collector.DB.CopyFrom(
		context.Background(),
		pgx.Identifier{"fiat_history"},
		[]string{"fiat_id", "date", "value"},
		pgx.CopyFromRows(insertFiat),
	)

	return
}
