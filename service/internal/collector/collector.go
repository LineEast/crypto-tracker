package collector

import (
	"context"
	"encoding/xml"
	"time"

	"github.com/LineEast/crypto-tracker/service/internal/database"
	"github.com/LineEast/crypto-tracker/service/internal/models"
	"github.com/LineEast/crypto-tracker/service/internal/server"

	json "github.com/goccy/go-json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valyala/fasthttp"
)

type (
	collector struct {
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

func New(db *pgxpool.Pool, config *Config) *collector {
	return &collector{
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

func (collector *collector) Run() (err error) {
	errs := make(chan error)

	fiatTicker := time.NewTicker(14 * time.Hour)
	go func() {
		for range fiatTicker.C {
			err := collector.fiat()
			if err != nil {
				errs <- err
			}
		}
	}()

	btcusdTticker := time.NewTicker(10 * time.Second)
	go func() {
		for range btcusdTticker.C {
			err := collector.btcusdt()
			if err != nil {
				errs <- err
			}
		}
	}()

	for err = range errs {
		if err != nil {
			break
		}
	}

	fiatTicker.Stop()
	btcusdTticker.Stop()

	return
}

func (collector *collector) fiat() (err error) {
	resp, err := server.MakeRequest(collector.FiatClient, server.XML)
	if err != nil {
		return
	}

	fiat := models.Fiat{}
	err = xml.Unmarshal(resp, &fiat)
	if err != nil {
		return
	}

	vl, err := database.SelectFiatMap(collector.DB)
	if err != nil {
		return
	}

	insertFiat := [][]interface{}{}
	for i := range fiat.Valute {
		id, ok := vl[fiat.Valute[i].FiatID]
		if ok {
			insertFiat = append(insertFiat, []interface{}{id, fiat.Date, fiat.Valute[i].Value})
		} else {
			err = database.InsertFiatRetID(collector.DB, &fiat.Valute[i])
			if err != nil {
				return
			}

			insertFiat = append(insertFiat, []interface{}{fiat.Valute[i].ID, fiat.Date, fiat.Valute[i].Value})
		}
	}

	err = database.InsertFiatCopyFrom(collector.DB, insertFiat)

	return
}

func (collector *collector) btcusdt() (err error) {
	resp, err := server.MakeRequest(collector.FiatClient, server.JSON)
	if err != nil {
		return
	}

	btcUsdt := models.BtcUsdt{}
	err = json.Unmarshal(resp, &btcUsdt)
	if err != nil {
		return
	}

	oldPrice, err := database.SelectAPBtcUsdt(collector.DB)
	if err != nil && err != pgx.ErrNoRows {
		return
	}

	// Если цена не изменилась, делать больше нечего
	if btcUsdt.AvaragePrice == oldPrice {
		return
	}

	// записываем новый курс в таблицу
	err = database.InsertBtcUsdt(collector.DB, &btcUsdt)
	if err != nil {
		return
	}

	// Перерасчет к фиатам
	vl, err := database.SelectFiatMap(collector.DB)
	if err != nil {
		return
	}
	valuteList := make([]models.Valute, 10)
	valute := models.Valute{}
	for _, v := range vl {
		valute.ID = v
		valuteList = append(valuteList, valute)
	}

	// На данный момент есть список всех фиатов, которые есть
	for i := range valuteList {
		// Собираем все курсы фиатных валют на данный момент
		err = database.SelectValFiat(collector.DB, &valuteList[i])

		// Делаем перерасчет и записываем в базу
		err = database.InsertFiatHistory(collector.DB, &btcUsdt, &valuteList[i])
		if err != nil {
			return
		}
	}

	return
}
