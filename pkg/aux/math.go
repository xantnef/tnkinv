package aux

import (
	"math"
	"time"

	log "github.com/sirupsen/logrus"
)

func Ratio2Perc(ratio float64) float64 {
	return (ratio - 1) * 100
}

func RatioAnnual(ratio float64, delta time.Duration) float64 {
	return math.Pow(ratio, 365/(delta.Hours()/24))
}

type payment struct {
	val  float64 // positive for investments
	date time.Time
}

type XirrCtx struct {
	payments []payment
}

func (ctx *XirrCtx) AddPayment(val float64, date time.Time) {
	// TODO is there a way to make it right efficient?
	ctx.payments = append(ctx.payments, payment{
		val:  val,
		date: date,
	})
}

func (ctx XirrCtx) Ratio(result float64, tn time.Time) float64 {
	epsilon := result / 100000.0
	if epsilon < 0.1 {
		epsilon = 0.1
	}

	if len(ctx.payments) == 0 {
		return 0
	}

	rate := 0.2
	rate_down, rate_up, have_rate_down, have_rate_up := 0.0, 0.0, false, false
	i := 0

	for i = 0; ; i++ {
		nv := -result
		for _, p := range ctx.payments {
			nv += p.val * math.Pow(1+rate, tn.Sub(p.date).Hours()/24.0/365.0)
		}
		log.Tracef("total %.2f e %f rate %f nv %f", result, epsilon, rate, nv)
		if nv < -epsilon {
			rate_down, have_rate_down = rate, true
			if have_rate_up {
				rate = (rate + rate_up) / 2
			} else {
				rate += 0.1
			}
		} else if nv > epsilon {
			rate_up, have_rate_up = rate, true
			if have_rate_down {
				rate = (rate + rate_down) / 2
			} else {
				rate -= 0.1
			}
		} else {
			break
		}
	}
	log.Infof("xirr took %d iterations, rate=%.2f, e=%.1f", i, rate, epsilon)
	return rate
}
