package aux

import (
	"math"
	"time"
)

func Ratio2Perc(ratio float64) float64 {
	return (ratio - 1) * 100
}

func RatioAnnual(ratio float64, delta time.Duration) float64 {
	return math.Pow(ratio, 365/(delta.Hours()/24))
}
