package portfolio

import (
	"fmt"
	"math"
	"time"

	log "github.com/sirupsen/logrus"

	"../candles"
	"../client"
	"../schema"
)

var beginning = time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)

type Portfolio struct {
	client *client.MyClient

	accs []string

	data struct {
		ops []schema.Operation
	}

	cc *candles.CandleCache

	instruments map[string]schema.Instrument // key=figi
	positions   map[string]*schema.PositionInfo

	accrued map[string]float64

	figisSorted []string

	balance schema.SectionedBalance
	alphas  schema.CurMap

	config struct {
		enableAccrued bool
		opsFile       string
		fictFile      string
	}
}

func NewPortfolio(c *client.MyClient, accs []string, opsFile, fictFile string) *Portfolio {
	p := &Portfolio{
		client: c,
		accs:   accs,

		instruments: make(map[string]schema.Instrument),
		positions:   make(map[string]*schema.PositionInfo),
		accrued:     make(map[string]float64),

		alphas: schema.NewCurMap(),
	}
	p.config.opsFile = opsFile
	p.config.fictFile = fictFile
	return p
}

// =============================================================================

func (p *Portfolio) addPosition(op schema.Operation) {
	if _, exists := p.positions[op.Figi]; exists {
		return
	}

	pinfo := &schema.PositionInfo{
		Ins: p.insByFigi(op.Figi),

		AccumulatedIncome: schema.NewCValue(0, op.Currency),
	}

	p.positions[op.Figi] = pinfo
}

// =============================================================================

func (p *Portfolio) processOperations(cb func(*schema.Balance, time.Time) bool) *schema.Balance {
	p.data.ops = p.getOperations(beginning)

	for _, op := range p.data.ops {
		if op.Status == "Done" && op.Figi != "" {
			p.addPosition(op)
		}
	}

	// gotta calc them repayments first, to be able to get correct prices
	// when calculating balances on the next iteration
	p.calculateRepayments()

	bal := schema.NewBalance()

	for _, op := range p.data.ops {
		if op.Status != "Done" {
			// cancelled declined etc
			// noone is interested in that
			continue
		}

		if !cb(bal, op.DateParsed) {
			break
		}

		log.Debugf("operation: %s", op)

		if op.Figi != "" {
			pinfo := p.positions[op.Figi]
			deal, isDeal := pinfo.AddOperation(op)
			if isDeal {
				bal.AddDeal(deal, pinfo.Ins.Figi)
			}
		}

		bal.AddOperation(op, p.cc.Xchgrate)

		log.Debugf(" [%s] %s at %s (%f) new balance: %f",
			op.OperationType, p.tryGetTicker(op.Figi),
			op.DateParsed.Format("2006/01/02"), op.Payment, bal.Assets["RUB"].Value)
	}

	for _, pinfo := range p.positions {
		if len(pinfo.Deals) == 0 {
			delete(p.positions, pinfo.Ins.Figi)
		}
	}

	return bal
}

// =============================================================================

func (p *Portfolio) getFullPrice(pinfo *schema.PositionInfo, t time.Time) float64 {
	return p.cc.Get(pinfo.Ins.Figi, t)*pinfo.RepaymentMultiplier(t) + p.getAccrued(pinfo, t)
}

func (p *Portfolio) openDealsSectionedBalance(time time.Time) schema.SectionedBalance {
	sb := schema.NewSectionedBalance()

	for _, pinfo := range p.positions {
		od, hasOd := pinfo.MakeOpenDeal(time,
			func() float64 {
				return p.getFullPrice(pinfo, time)
			})

		if !hasOd || pinfo.Ins.Figi == schema.FigiUSD {
			continue
		}

		log.Debugf("open deal %s %s %s", pinfo.Ins.Figi, pinfo.Ins.Ticker, od)

		sb.AddDeal(od, pinfo.Ins.Figi, pinfo.Ins.Section)
	}

	return sb
}

func (p *Portfolio) openDealsBalance(time time.Time) *schema.Balance {
	sb := p.openDealsSectionedBalance(time)
	log.Debugf("current asset balance: %s", sb.Total)
	return sb.Total
}

func (p *Portfolio) Collect(at time.Time) {
	p.collectAccrued()

	p.cc = candles.NewCandleCache(p.client)

	cash := p.processOperations(func(bal *schema.Balance, opTime time.Time) bool {
		return opTime.Before(at)
	})

	p.balance = p.openDealsSectionedBalance(at)
	p.balance.Total.Add(*cash)

	for _, pinfo := range p.positions {
		alpha := p.makePortionYields(pinfo)
		p.alphas.Add(alpha)
	}

	p.calcAllAssets(p.balance, p.alphas, at)
}

// =============================================================================

func (p *Portfolio) ListDeals(start, end time.Time) {
	empty := true

	p.data.ops = p.getOperations(start)

	deals := schema.NewBalance()
	comms := schema.NewBalance()
	for _, op := range p.data.ops {
		if op.DateParsed.After(end) {
			break
		}

		if op.Status != "Done" {
			// cancelled declined etc
			// noone is interested in that
			continue
		}

		if op.Figi != "" {
			op.Ticker = p.insByFigi(op.Figi).Ticker
		}
		fmt.Printf("%s\n", op.StringPretty())

		// exploit those balance maps for totals
		if op.IsTrading() {
			deals.Assets[op.Currency].Value += math.Abs(op.Payment)
			empty = false
		} else if op.OperationType == "ServiceCommission" || op.OperationType == "BrokerCommission" {
			comms.Assets[op.Currency].Value += math.Abs(op.Payment)
			empty = false
		}
	}

	if empty {
		return
	}

	fmt.Printf(" - Total deals:\n")
	for _, c := range schema.CurrenciesOrdered {
		if deals.Assets[c].Value != 0 {
			fmt.Printf("\t %s\n", deals.Assets[c])
		}
	}
	fmt.Printf("   commissions:\n")
	for _, c := range schema.CurrenciesOrdered {
		if comms.Assets[c].Value != 0 {
			fmt.Printf("\t %s\n", comms.Assets[c])
		}
	}

	usdrate := p.client.RequestCurrentPrice(schema.FigiUSD)
	fmt.Printf("   percentage: %.2f%%\n", comms.CalcAllAssets(usdrate, 0)/deals.CalcAllAssets(usdrate, 0)*100)
}

// =============================================================================

func (p *Portfolio) calcAllAssets(sb schema.SectionedBalance, alphas schema.CurMap, t time.Time) {
	usd, eur := p.cc.Get(schema.FigiUSD, t), 0.0

	sb.CalcAllAssets(usd, eur)

	if alphas != nil {
		alphas.CalcAll(usd, eur)
	}
}

func (p *Portfolio) summarize( /* const */ bal schema.Balance, t time.Time, format string) {
	obal := p.openDealsSectionedBalance(t)
	obal.Total.Add(bal)

	p.calcAllAssets(obal, nil, t)

	obal.Print(t.Format("2006/01/02"), format)
}

func (p *Portfolio) ListBalances(start time.Time, period, format string) {
	p.cc = candles.NewCandleCache(p.client).WithPeriod(start, period)

	candleTimes := p.cc.ListTimes()

	cidx := 0
	num := len(candleTimes)

	if num == 0 {
		log.Debug("No data for this period")
		return
	}

	schema.PrintBalanceHead(format)

	bal := p.processOperations(func(bal *schema.Balance, opTime time.Time) bool {

		// process all candles before opTime

		for ; cidx < num; cidx += 1 {
			nextTime := candleTimes[cidx]
			if opTime.Before(nextTime) {
				break
			}
			p.summarize(*bal, nextTime, format)
		}

		return true
	})

	log.Debugf("cash balance: %s", bal.Assets)

	// process all candles after the last operation

	for ; cidx < num; cidx += 1 {
		nextTime := candleTimes[cidx]
		p.summarize(*bal, nextTime, format)
	}
}
