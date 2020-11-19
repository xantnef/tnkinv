package candles

import (
	"errors"
	"time"

	log "github.com/sirupsen/logrus"

	"../schema"
)

func (cc *CandleCache) WithPeriod(start time.Time, period string) *CandleCache {
	cc.start = start
	cc.period = period
	cc.pcache = make(candleMap)
	return cc
}

func strToDuration(s string) time.Duration {
	switch s {
	case "day":
		return 24 * time.Hour
	case "week":
		return 7 * strToDuration("day")
	case "month":
		return strToDuration("year") / 12
	case "year":
		return 365 * strToDuration("day")
	default:
		log.Fatalf("Unknown period %s", s)
		return 0
	}
}

func (cc *CandleCache) doFetchPeriod(figi string) (clist []candle) {
	defer print(figi, clist)

	if cc.period == "day" {
		return cc.fetchDaily(figi, cc.start, time.Now())
	}

	t1 := cc.start.Add(-strToDuration(cc.period))
	t2 := time.Now()

	pcandles := cc.client.RequestCandles(figi, t1, t2, cc.period).Payload.Candles

	if len(pcandles) < 1 {
		log.Fatalf("No candles for period %s - %s (%s)", t1, t2, cc.start)
	}

	for _, p := range pcandles {
		log.Debugf("<c> %s %s %.2f %.2f", figi, p.Time, p.O, p.C)
	}

	for i := 1; i < len(pcandles); i++ {
		date, err := time.Parse(time.RFC3339, pcandles[i].Time)
		if err != nil {
			log.Fatalf("failed to parse time: %v (%s)", pcandles[i], err)
		}

		clist = append(clist, candle{
			time:  date,
			price: pcandles[i-1].C,
		})
	}

	clist = append(clist, candle{
		time:  t2,
		price: pcandles[len(pcandles)-1].C,
	})

	return clist
}

func (cc *CandleCache) fetchPeriod(figi string) {
	if _, exist := cc.pcache[figi]; exist {
		return
	}

	pcandles := cc.doFetchPeriod(figi)
	cc.pcache[figi] = sortCandles(append(cc.pcache[figi], pcandles...))
}

func (cc *CandleCache) getPeriodic(figi string, t time.Time) (float64, error) {
	if cc.period == "" {
		return 0, errors.New("no period")
	}

	cc.fetchPeriod(figi)

	return cc.pcache.tryFind(figi, t)
}

func (cc *CandleCache) ListTimes() (times []time.Time) {
	if cc.period == "" {
		log.Fatal("no cache period")
	}

	cc.fetchPeriod(schema.FigiUSD)
	for _, c := range cc.pcache[schema.FigiUSD] {
		times = append(times, c.time)
	}
	return times
}
