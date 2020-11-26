package aux

import (
	"testing"
	"time"
)

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func TestXirr(t *testing.T) {
	var ctx XirrCtx
	ctx.AddPayment(100, date(2002, 1, 1))
	ctx.AddPayment(100, date(2003, 1, 1))
	rate := ctx.Ratio(231, date(2004, 1, 1))
	if rate != 0.1 {
		t.Errorf("xiir() = %f, exp 0.1", rate)
	}
}
