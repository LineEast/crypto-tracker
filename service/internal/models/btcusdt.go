package models

import "time"

type (
	BtcUsdt struct {
		ID           string
		Time         time.Time `json:"time"`
		AvaragePrice float32   `json:"avaragePrice"` // на тестовое задание будет float, но на продакшене лучше использовать string/int
	}
)
