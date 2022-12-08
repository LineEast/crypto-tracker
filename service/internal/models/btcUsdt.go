package models

type (
	BtcUsdt struct {
		ID           string
		Time         int64   `json:"time"`
		AvaragePrice float32 `json:"avaragePrice"` // на тестовое задание будет float, но на продакшене лучше использовать string/int
	}
)
