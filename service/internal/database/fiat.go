package database

import (
	"context"

	"github.com/LineEast/crypto-tracker/service/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

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
