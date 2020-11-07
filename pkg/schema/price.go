package schema

import (
	"time"
)

type Pricef0 func() float64
type Pricef1 func(figi string) float64
type Pricef2 func(figi string, t time.Time) float64

func PriceCurry0(f1 Pricef1, figi string) Pricef0 {
	return func() float64 {
		return f1(figi)
	}
}
