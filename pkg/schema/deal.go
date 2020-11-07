package schema

import (
	"fmt"
	"time"
)

type Deal struct {
	Date       time.Time
	Price      CValue
	Quantity   int
	Accrued    float64 // aka NKD
	Commission float64
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
