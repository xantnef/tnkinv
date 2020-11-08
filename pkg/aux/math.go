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

func AdjustDate(avgT, newT time.Time, total, value float64) time.Time {
	// TODO think again is this correct?
	mult := value / total
	biasDays := int(math.Round(newT.Sub(avgT).Hours() * mult / 24))
	return avgT.AddDate(0, 0, biasDays)
}
