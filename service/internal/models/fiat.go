package models

import "time"

type (
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

	FiatHistory struct {
		Date   time.Time `json:"date"`
		Valute Valute    `json:"valute"`
	}
)
