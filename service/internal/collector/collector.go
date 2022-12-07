package collector

import (
	"context"
	"encoding/xml"
	"time"

	json "github.com/goccy/go-json"
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
		Value  float32 `xml:"Value"`
	}

	BtcUsdt struct {
		ID           string
		Time         time.Time `json:"time"`
		AvaragePrice float32   `json:"avaragePrice"` // на тестовое задание будет float, но на продакшене лучше использовать string/int
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
	req.Header.SetMethod(fasthttp.MethodGet)
	req.Header.Set(fasthttp.HeaderContentType, "application/xml")
	req.SetRequestURI(collector.config.FiatEndPoint)
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
		id, ok := vl[fiat.Valute[i].FiatID]
		if ok {
			insertFiat = append(insertFiat, []interface{}{id, fiat.Date, fiat.Valute[i].Value})
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

func (collector *Collector) btcusdt() (err error) {
	req := fasthttp.AcquireRequest()
	req.Header.SetMethod(fasthttp.MethodGet)
	req.Header.Set(fasthttp.HeaderContentType, "application/json")
	req.SetRequestURI(collector.config.CryptoEndPoint)
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)

	err = collector.CryptoClient.Do(req, resp)
	if err != nil {
		return
	}

	btcUsdt := BtcUsdt{}
	err = json.Unmarshal(resp.Body(), &btcUsdt)
	if err != nil {
		return
	}

	// Получаем предыдущую цену, если она есть
	row, err := collector.DB.Query(
		context.Background(),
		"select avarage_price from btc_usdt order by date desc limit 1",
	)
	row.Close()
	if err != nil {
		return
	}

	var oldPrice float32
	if row.Next() {
		err = row.Scan(&oldPrice)
		if err != nil {
			return
		}
	}
	// ===============

	if btcUsdt.AvaragePrice == oldPrice {
		return
	}

	// записываем новый курс в таблицу
	err = collector.DB.QueryRow(
		context.Background(),
		"insert into btc_usdt (date, avarage_price) values ($1, $2) returning id",
		btcUsdt.Time, btcUsdt.AvaragePrice,
	).Scan(&btcUsdt.ID)
	if err != nil {
		return
	}

	// Перерасчет к фиатам
	valuteList := make([]Valute, 10)
	valute := Valute{}

	rows, err := collector.DB.Query(
		context.Background(),
		"select id, code from fiat",
	)
	defer rows.Close()
	if err != nil {
		return err
	}

	for rows.Next() {
		err = rows.Scan(&valute.ID, &valute.Code)
		if err != nil {
			return err
		}

		valuteList = append(valuteList, valute)
	}

	// На данный момент есть список всех фиатов, которые есть
	for i := range valuteList {
		// Собираем все курсы фиатных валют на данный момент
		err = collector.DB.QueryRow(
			context.Background(),
			"select value from fiat_history where fiat_id=$1 order by date desc limit 1",
			valuteList[i].ID,
		).Scan(&valuteList[i].Value)
		if err != nil {
			return
		}

		// Делаем перерасчет
		_, err = collector.DB.Exec(
			context.Background(),
			"insert into btc_usdt_history (btc_usdt_id, fiat_id, value) values ($1, $2, $3)",
			btcUsdt.ID, valuteList[i].ID, btcUsdt.AvaragePrice*valuteList[i].Value,
		)
		if err != nil {
			return
		}
	}

	// (BTC -> USD) -> RUB -> в базу

	return
}

func (bu *BtcUsdt) recalculation() (err error) {

	return
}
