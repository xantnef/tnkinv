package candles

import (
	"errors"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"

	"../client"
	"../schema"
)

// TODO
// Candles for amortized bonds are quire fucked up.
// They show the old values as if it was now.
// e.g.
//   1. price = 1000
//   2. grew to 1100
//   3. amortized to 880
// candle for the time point (1) is going to show 800.

type candleMap map[string][]candle // key=figi

type candle struct {
	price float64
	time  time.Time
}

type CandleCache struct {
	client *client.MyClient
	cache  candleMap

	start  time.Time
	period string
	pcache candleMap
}

func NewCandleCache(c *client.MyClient) *CandleCache {
	return &CandleCache{
		client: c,
		cache:  make(candleMap),
	}
}

func normalize(t time.Time) time.Time {
	// normalize the time a bit
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	// we wont catch all the holidays here, but it covers 95%
	wday := t.Weekday()
	if wday == time.Saturday {
		t = t.Add(-1 * 24 * time.Hour)
	} else if wday == time.Sunday {
		t = t.Add(-2 * 24 * time.Hour)
	}

	return t
}

func (cc *CandleCache) fetchDay(figi string, t time.Time) (clist []candle) {
	defer print(figi, clist)

	clist = cc.fetchDaily(figi, t, t.Add(24*time.Hour))

	// idx = first element after or equal to day @t
	idx := sort.Search(len(clist), func(i int) bool {
		el := clist[i].time
		return el.Year() > t.Year() ||
			el.Year() == t.Year() && el.YearDay() >= t.YearDay()
	})

	if idx < len(clist) {
		el := clist[idx].time
		if el.Year() == t.Year() && el.YearDay() == t.YearDay() {
			return clist
		}
	}

	if idx == 0 {
		// all candles are after
		log.Fatalf("unexpected candle list: %s %v %s", figi, clist, t)
	}

	clist = append(clist, candle{
		time:  t,
		price: clist[idx-1].price,
	})

	return clist
}

func (cc *CandleCache) fetchDaily(figi string, t1, t2 time.Time) (clist []candle) {
	t1 = normalize(t1)

	pcandles := cc.client.RequestCandles(figi, t1, t2, "day").Payload.Candles

	if len(pcandles) < 1 {
		log.Debugf("No candles for period %s- %s", t1, t2)
		return cc.fetchDaily(figi, t1.Add(-24*time.Hour), t2)
	}

	for _, p := range pcandles {
		log.Debugf("<c> %s %s %.2f %.2f", figi, p.Time, p.O, p.C)
	}

	for _, p := range pcandles {
		date, err := time.Parse(time.RFC3339, p.Time)
		if err != nil {
			log.Fatalf("failed to parse time: %v (%s)", p, err)
		}

		clist = append(clist, candle{
			time:  date,
			price: p.C,
		})
	}

	return clist
}

func print(figi string, clist []candle) {
	for _, p := range clist {
		log.Debugf("price %s(%s) = %.2f", figi, p.time, p.price)
	}
}

func sortCandles(pcandles []candle) []candle {
	sort.Slice(pcandles, func(i, j int) bool {
		return pcandles[i].time.Before(pcandles[j].time)
	})
	return pcandles
}

func (cm candleMap) tryFind(figi string, t time.Time) (float64, error) {
	logAndRet := func(c candle) (float64, error) {
		log.Debugf("price(%s on %s) = %s : %f",
			figi, t, c.time, c.price)
		return c.price, nil
	}

	pcandles, exist := cm[figi]
	if !exist {
		return 0, errors.New("no cache")
	}

	// idx = first element after date X
	idx := sort.Search(len(pcandles), func(i int) bool {
		el := pcandles[i].time
		return el.Year() > t.Year() ||
			el.Year() == t.Year() && el.YearDay() >= t.YearDay()
	})

	if idx == len(pcandles) {
		// all candles are before
		return 0, errors.New("date is too late")
	}

	el := pcandles[idx].time
	if el.Year() == t.Year() && el.YearDay() == t.YearDay() {
		return logAndRet(pcandles[idx])
	}

	return 0, errors.New("exact date not found")
}

func (cm candleMap) find(figi string, t time.Time) float64 {
	price, err := cm.tryFind(figi, t)
	if err != nil {
		log.Fatalf("No candle %s %s: %s", figi, t, err)
	}
	return price
}

func (cc *CandleCache) Get(figi string, t time.Time) float64 {
	p, err := cc.getPeriodic(figi, t)
	if err == nil {
		return p
	}

	price, err := cc.cache.tryFind(figi, t)
	if err == nil {
		return price
	}

	pcandles := cc.fetchDay(figi, t)

	cc.cache[figi] = sortCandles(append(cc.cache[figi], pcandles...))

	return cc.cache.find(figi, t)
}

func (cc *CandleCache) PriceFigi(t time.Time) schema.PriceFigi {
	return func(figi string) float64 {
		return cc.Get(figi, t)
	}
}

func (cc *CandleCache) Xchgrate(curr_from, curr_to string, t time.Time) float64 {
	if xf, ok := map[string]func() float64{
		"RUB" + "RUB": func() float64 { return 1 },
		"RUB" + "USD": func() float64 { return 1 / cc.Get(schema.FigiUSD, t) },
		"USD" + "USD": func() float64 { return 1 },
		"USD" + "RUB": func() float64 { return cc.Get(schema.FigiUSD, t) },
	}[curr_from+curr_to]; ok {
		return xf()
	}
	log.Fatalf("unknown conversion %s->%s", curr_from, curr_to)
	return 0
}

func (cc *CandleCache) GetInCurrency(ins schema.Instrument, curr string, t time.Time) float64 {
	price := cc.Get(ins.Figi, t) * cc.Xchgrate(ins.Currency, curr, t)
	log.Debugf("%s at %s costs %f", ins.Ticker, t, price)
	return price
}
