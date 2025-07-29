package models

import (
	"github.com/shopspring/decimal"
	"time"
)

type Transaction struct {
	Amount decimal.Decimal
	From   string
	To     string
	Time   time.Time
}
