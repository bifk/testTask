package models

import (
	"github.com/shopspring/decimal"
)

type Wallet struct {
	Address string
	Amount  decimal.Decimal
}
