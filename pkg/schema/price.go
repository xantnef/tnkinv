package schema

import (
	"time"
)

type PriceF func() float64
type PriceFigi func(figi string) float64
type PriceFigiAt func(figi string, t time.Time) float64
type PriceAt func(time.Time) float64

func PriceCurry0(f1 PriceFigi, figi string) PriceF {
	return func() float64 {
		return f1(figi)
	}
}
