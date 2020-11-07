package candles

import (
	"errors"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"

	"../aux"
	"../client"
	"../schema"
)

type candle struct {
	price float64
	time  time.Time
}

type CandleCache struct {
	client *client.MyClient
	start  time.Time
	period string

	periodFetched aux.List            // key=figi
	cache         map[string][]candle // key=figi
}

func NewCandleCache(c *client.MyClient) *CandleCache {
	return &CandleCache{
		client: c,

		periodFetched: aux.NewList(),
		cache:         make(map[string][]candle),
	}
}

func (cc *CandleCache) WithPeriod(start time.Time, period string) *CandleCache {
	cc.start = start
	cc.period = period
	return cc
}

func strToDuration(s string) time.Duration {
	switch s {
	case "day":
		return 24 * time.Hour
	case "week":
		return 7 * 24 * time.Hour
	case "month":
		return 365 * 24 * time.Hour / 12
	case "year":
		return 365 * 24 * time.Hour
	default:
		log.Fatalf("Unknown period %s", s)
		return 0
	}
}

func (cc *CandleCache) fetchCandles(figi string, start, end time.Time, period string) (clist []candle) {
	adjustment := 2 * strToDuration(period)
	resp := cc.client.RequestCandles(figi, start.Add(-adjustment), end, period)
	pcandles := resp.Payload.Candles

	// TODO
	// Candles for amortized bonds are quire fucked up.
	// They show the old values as if it was now.
	// e.g.
	//   1. price = 1000
	//   2. grew to 1100
	//   3. amortized to 880
	// candle for the time point (1) is going to show 800.

	if len(pcandles) < 2 {
		log.Debugf("No candles for period %s - %s", start, end)
		return cc.fetchCandles(figi, start.Add(-24*time.Hour), end, period)
	}

	for _, p := range pcandles {
		log.Debugf("<c> %s %s %.2f %.2f", figi, p.Time, p.O, p.C)
	}

	for i := 1; i < len(pcandles); i++ {
		var err error
		var c candle

		c.price = pcandles[i-1].C
		c.time, err = time.Parse(time.RFC3339, pcandles[i].Time)
		if err != nil {
			log.Fatal("failed to parse time: %v (%s)", pcandles[i], err)
		}

		log.Debugf("new candle %s=%f ", c.time, c.price)

		clist = append(clist, c)
	}

	clist = append(clist, candle{
		time:  end,
		price: pcandles[len(pcandles)-1].C,
	})

	return clist
}

func sortCandles(pcandles []candle) []candle {
	sort.Slice(pcandles, func(i, j int) bool {
		return pcandles[i].time.Before(pcandles[j].time)
	})
	return pcandles
}

func (cc *CandleCache) fetchPeriod(figi string) {
	if cc.periodFetched.Has(figi) {
		return
	}
	cc.periodFetched.Add(figi)

	now := time.Now()

	pcandles := cc.fetchCandles(figi, cc.start, now, cc.period)
	cc.cache[figi] = sortCandles(append(cc.cache[figi], pcandles...))
}

func (cc *CandleCache) findInCache(figi string, t time.Time, exact bool) (float64, error) {
	logAndRet := func(c candle) (float64, error) {
		log.Debugf("price(%s on %s) = %s : %f",
			figi, t, c.time, c.price)
		return c.price, nil
	}

	pcandles, exist := cc.cache[figi]
	if !exist {
		return 0, errors.New("no cache")
	}

	// idx = first element after date X
	idx := sort.Search(len(pcandles), func(i int) bool {
		el := pcandles[i].time
		return el.Year() > t.Year() ||
			el.Year() == t.Year() && el.YearDay() >= t.YearDay()
	})

	if idx < len(pcandles) {
		el := pcandles[idx].time
		if el.Year() == t.Year() && el.YearDay() == t.YearDay() {
			return logAndRet(pcandles[idx])
		}
	}

	if exact {
		return 0, errors.New("exact date not found")
	}

	if idx == 0 {
		// all candles are after
		return 0, errors.New("date is too early")
	}

	if idx == len(pcandles) {
		// all candles are before
		return 0, errors.New("date is too late")
	}

	if t.Sub(pcandles[idx-1].time) < pcandles[idx].time.Sub(t) {
		return logAndRet(pcandles[idx-1])
	} else {
		return logAndRet(pcandles[idx])
	}
}

func (cc *CandleCache) find(figi string, t time.Time) float64 {
	price, err := cc.findInCache(figi, t, false)
	if err != nil {
		log.Fatalf("Couldnt get candles? %s %s: %s", figi, t, err)
	}
	return price
}

func (cc *CandleCache) findExactDate(figi string, t time.Time) (float64, error) {
	return cc.findInCache(figi, t, true)
}

func (cc *CandleCache) Get(figi string, t time.Time) float64 {
	if cc.period != "" {
		cc.fetchPeriod(figi)
		return cc.find(figi, t)
	}
	return cc.GetOnDay(figi, t)
}

func (cc *CandleCache) GetOnDay(figi string, t time.Time) float64 {
	price, err := cc.findExactDate(figi, t)
	if err == nil {
		return price
	}

	// normalize the time a bit
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	// we wont catch all the holidays here, but it covers 95%
	wday := start.Weekday()
	if wday == time.Saturday {
		start = start.Add(-1 * 24 * time.Hour)
	} else if wday == time.Sunday {
		start = start.Add(-2 * 24 * time.Hour)
	}

	pcandles := cc.fetchCandles(figi, start, start.Add(24*time.Hour), "day")

	// might be a duplicate. Nothing critical
	cc.cache[figi] = sortCandles(append(cc.cache[figi], pcandles...))

	return cc.find(figi, t)
}

func (cc *CandleCache) ListTimes() (times []time.Time) {
	cc.fetchPeriod(schema.FigiUSD)
	for _, c := range cc.cache[schema.FigiUSD] {
		times = append(times, c.time)
	}
	return
}

func (cc *CandleCache) Pricef1(t time.Time) schema.Pricef1 {
	return func(figi string) float64 {
		return cc.Get(figi, t)
	}
}
