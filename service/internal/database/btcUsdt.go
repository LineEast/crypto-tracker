package database

import (
	"context"

	"github.com/LineEast/crypto-tracker/service/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SelectBtcUsdtLast(db *pgxpool.Pool, btcusdt *models.BtcUsdt) (err error) {
	row := db.QueryRow(
		context.Background(),
		"select date, avarage_price from btc_usdt order by date desc limit 1",
	)

	err = row.Scan(&btcusdt.Time, &btcusdt.AvaragePrice)
	if err != nil {
		return
	}

	return
}

func SelectBtcUsdt(db *pgxpool.Pool) (btcusdtList []models.BtcUsdt, err error) {
	rows, err := db.Query(
		context.Background(),
		"select date, avarage_price from btc_usdt order by date",
	)
	if err != nil {
		return
	}
	defer rows.Close()

	btcusdt := models.BtcUsdt{}
	for rows.Next() {
		err = rows.Scan(&btcusdt.Time, &btcusdt.AvaragePrice)
		if err != nil {
			return
		}

		btcusdtList = append(btcusdtList, btcusdt)
	}
	if err != nil {
		return
	}

	return
}

// Получаем предыдущую цену, если она есть
func SelectAPBtcUsdt(db *pgxpool.Pool) (oldPrice float32, err error) {
	row := db.QueryRow(
		context.Background(),
		"select avarage_price from btc_usdt order by date desc limit 1",
	)

	err = row.Scan(&oldPrice)
	if err != nil {
		return
	}

	return
}

func InsertBtcUsdt(db *pgxpool.Pool, s *models.BtcUsdt) (err error) {
	err = db.QueryRow(
		context.Background(),
		"insert into btc_usdt (date, avarage_price) values ($1, $2) returning id",
		s.Time, s.AvaragePrice,
	).Scan(&s.ID)
	if err != nil {
		return
	}

	return
}

func InsertFiatHistory(db *pgxpool.Pool, btcUsdt *models.BtcUsdt, valute *models.Valute) (err error) {
	_, err = db.Exec(
		context.Background(),
		"insert into btc_usdt_history (btc_usdt_id, fiat_id, value) values ($1, $2, $3)",
		btcUsdt.ID, valute.ID, btcUsdt.AvaragePrice*valute.Value,
	)
	if err != nil {
		return
	}

	return
}
