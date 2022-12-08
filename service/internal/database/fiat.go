package database

import (
	"context"

	"github.com/LineEast/crypto-tracker/service/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SelectFiatHistoryLast(db *pgxpool.Pool) (fiatList []models.FiatHistory, err error) {
	vl, err := SelectFiat(db)
	if err != nil {
		return
	}

	fiat := models.FiatHistory{}
	for i := range vl {
		err = db.QueryRow(
			context.Background(),
			"select date, value from fiat_history where fiat_id=$1 order by date limit 1",
			vl[i].ID,
		).Scan(&fiat.Date, &fiat.Valute.Value)
		if err != nil {
			return
		}

		fiat.Valute.Code = vl[i].Code
		fiat.Valute.Name = vl[i].Name
		fiat.Valute.FiatID = vl[i].FiatID

		fiatList = append(fiatList, fiat)
	}

	return
}

func SelectFiat(db *pgxpool.Pool) (vl []models.Valute, err error) {
	rows, err := db.Query(
		context.Background(),
		"select id, fiat_id, name, code from fiat",
	)
	if err != nil {
		return
	}
	defer rows.Close()

	valute := models.Valute{}
	for rows.Next() {
		err = rows.Scan(&valute.ID, &valute.FiatID, &valute.Name, &valute.Code)
		if err != nil {
			return
		}

		vl = append(vl, valute)
	}

	return
}

// возвращает map[айди_в_базе]айди_самой_валюты
func SelectFiatMap(db *pgxpool.Pool) (vl map[string]string, err error) {
	rows, err := db.Query(
		context.Background(),
		"select id, fiat_id from fiat",
	)
	if err != nil {
		return
	}
	defer rows.Close()

	valute := models.Valute{}
	for rows.Next() {
		err = rows.Scan(&valute.ID, &valute.FiatID)
		if err != nil {
			return
		}

		vl[valute.FiatID] = valute.ID
	}

	return
}

func SelectValFiat(db *pgxpool.Pool, v *models.Valute) (err error) {
	err = db.QueryRow(
		context.Background(),
		"select value from fiat_history where fiat_id=$1 order by date desc limit 1",
		v.ID,
	).Scan(&v.Value)
	if err != nil {
		return
	}

	return
}

func InsertFiatRetID(db *pgxpool.Pool, v *models.Valute) (err error) {
	err = db.QueryRow(
		context.Background(),
		"insert into fiat (fiat_id, name, code) values ($1, $2, $3) returning id",
		v.FiatID, v.Name, v.Code,
	).Scan(&v.ID)
	if err != nil {
		return
	}

	return
}

func InsertFiatCopyFrom(db *pgxpool.Pool, insertFiat [][]interface{}) (err error) {
	_, err = db.CopyFrom(
		context.Background(),
		pgx.Identifier{"fiat_history"},
		[]string{"fiat_id", "date", "value"},
		pgx.CopyFromRows(insertFiat),
	)

	return
}
