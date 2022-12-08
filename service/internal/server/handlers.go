package server

import (
	"github.com/LineEast/crypto-tracker/service/internal/database"
	"github.com/LineEast/crypto-tracker/service/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valyala/fasthttp"
)

type (
	btcUsdt struct {
		Total   uint             `json:"Total"`
		History []models.BtcUsdt `json:"history"`
	}

	fiat struct {
		Total   uint                 `json:"total"`
		History []models.FiatHistory `json:"history"`
	}
)

// запрос должен выводить последнее (текущее) значение
func btusdtGET(db *pgxpool.Pool) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		btcUsdt := models.BtcUsdt{}
		err := database.SelectBtcUsdtLast(db, &btcUsdt)
		if err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
			return
		}
		resp := models.BtcUsdt{
			Time:         btcUsdt.Time,
			AvaragePrice: btcUsdt.AvaragePrice,
		}

		makeResponse(ctx, resp)
	}
}

// историю с фильтрами по дате и времени и пагинацией
func btusdtPOST(db *pgxpool.Pool) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		btcUsdt := btcUsdt{}
		btcUsdtList, err := database.SelectBtcUsdt(db)
		if err != nil {
			return
		}

		for i := range btcUsdtList {
			btcUsdt.History = append(btcUsdt.History, btcUsdtList[i])
		}
	}
}

// запрос отображает последние (текущие) курсы фиатных валют по отношению к рублю
func currenciesGET(db *pgxpool.Pool) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		fiatHistory, err := database.SelectFiatHistoryLast(db)
		if err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
			return
		}

		makeResponse(ctx, fiatHistory)
	}
}

// запрос возвращает историю изменения с фильтрами по дате и валюте и пагинацией
func currenciesPOST(db *pgxpool.Pool) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {

	}
}
