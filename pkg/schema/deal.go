package schema

import (
	"fmt"
	"time"
)

type Deal struct {
	Date       time.Time
	Price      CValue
	Quantity   int     // positive for Buy
	Accrued    float64 // aka NKD
	Commission float64 // negative
}

func (deal Deal) IsBuy() bool {
	return deal.Quantity > 0
}

// positive for Buy
func (deal Deal) Expense() float64 {
	return deal.Value() - deal.Commission
}

// positive for Sell
func (deal Deal) Profit() float64 {
	return -deal.Expense()
}

func (deal Deal) String() string {
	return fmt.Sprintf(
		"%s: (%.2f x %d) = %s",
		deal.Date.Format("2006/01/02"),
		deal.Price.Value,
		deal.Quantity,
		deal.CValue())
}

func (deal Deal) Value() float64 {
	return deal.Price.Value*float64(deal.Quantity) + deal.Accrued
}

func (deal Deal) CValue() CValue {
	cv := deal.Price.Mult(float64(deal.Quantity))
	cv.Value += deal.Accrued
	return cv
}
