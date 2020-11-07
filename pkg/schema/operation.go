package schema

import (
	"fmt"

	"../aux"
)

func (op Operation) StringPretty() string {
	shortTick := op.Ticker
	if op.Figi == FigiUSD {
		shortTick = "USD"
	}

	return fmt.Sprintf(
		"%s: %-17s %-4s (%-7.2f x %-3d) = %s %-9.2f",
		op.DateParsed.Format("2006/01/02"), op.OperationType, shortTick, op.Price, op.Quantity,
		op.Currency, op.Payment)
}

func (op Operation) IsTrading() bool {
	return aux.IsIn(op.OperationType, "Buy", "BuyCard", "Sell")
}

func (op Operation) IsPayment() bool {
	// TODO PartRepayment doesnt really belong here.
	// make it a pseudo-deal?
	return aux.IsIn(op.OperationType,
		"Dividend",
		"TaxDividend",
		"Coupon",
		"TaxCoupon",
		"PartRepayment",
	)
}
