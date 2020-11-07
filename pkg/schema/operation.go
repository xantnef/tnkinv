package schema

import (
	"fmt"
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
	return op.OperationType == "Buy" || op.OperationType == "BuyCard" ||
		op.OperationType == "Sell"
}

func (op Operation) IsPayment() bool {
	// TODO PartRepayment doesnt really belong here.
	// make it a pseudo-deal?
	return op.OperationType == "Dividend" ||
		op.OperationType == "TaxDividend" ||
		op.OperationType == "Coupon" ||
		op.OperationType == "TaxCoupon" ||
		op.OperationType == "PartRepayment"
}
