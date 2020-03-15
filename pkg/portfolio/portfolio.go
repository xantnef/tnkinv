package portfolio

import (
	"fmt"
	"math"
	"sort"
	"sync"
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

	tickers   map[string]string
	positions map[string]*schema.PositionInfo

	accrued map[string]float64

	figisSorted []string

	cash, funds, bonds, stocks, totals *schema.Balance // may be nil!

	config struct {
		enableAccrued bool
	}
}

func NewPortfolio(c *client.MyClient, accs []string) *Portfolio {
	return &Portfolio{
		client:    c,
		accs:      accs,
		tickers:   make(map[string]string),
		positions: make(map[string]*schema.PositionInfo),
		accrued:   make(map[string]float64),
	}
}

// =============================================================================

func (p *Portfolio) getTicker(figi string) string {
	ticker, ok := p.tickers[figi]
	if !ok {
		ticker = p.client.RequestTicker(figi)
		p.tickers[figi] = ticker
	}
	return ticker
}

func (p *Portfolio) getFigi(ticker string) string {
	for figi, tick := range p.tickers {
		if tick == ticker {
			return figi
		}
	}

	figi := p.client.RequestFigi(ticker)
	p.tickers[figi] = ticker
	return figi
}

// =============================================================================

func (p *Portfolio) processPortfolio() {
	for _, acc := range p.accs {
		pfResp := p.client.RequestPortfolio(acc)
		for _, pos := range pfResp.Payload.Positions {
			p.tickers[pos.Figi] = pos.Ticker
			p.accrued[pos.Figi] = pos.AveragePositionPrice.Value - pos.AveragePositionPriceNoNkd.Value
		}
	}
}

// =============================================================================

func (p *Portfolio) preprocessOperations(start time.Time) {
	var ops []schema.Operation

	for _, acc := range p.accs {
		resp := p.client.RequestOperations(start, acc)
		ops = append(ops, resp.Payload.Operations...)
	}

	for i := range ops {
		var err error
		ops[i].DateParsed, err = time.Parse(time.RFC3339, ops[i].Date)
		if err != nil {
			log.Fatal("Failed to parse time: %v", err)
		}
	}

	sort.Slice(ops, func(i, j int) bool {
		return ops[i].DateParsed.Before(ops[j].DateParsed)
	})

	p.data.ops = ops
}

// =============================================================================

func (p *Portfolio) processOperation(op schema.Operation) (deal *schema.Deal) {
	log.Debugf("%s", op)

	if op.Figi == "" {
		// payins, service commissions
		return
	}

	pinfo := p.positions[op.Figi]
	if pinfo == nil {
		pinfo = &schema.PositionInfo{
			Figi:   op.Figi,
			Ticker: p.getTicker(op.Figi),
			Type:   op.InstrumentType,

			AccumulatedIncome: schema.NewCValue(0, op.Currency),
		}

		// catch unhandled
		{
			m := map[string]bool{
				schema.InsTypeEtf:      true,
				schema.InsTypeStock:    true,
				schema.InsTypeBond:     true,
				schema.InsTypeCurrency: true,
			}
			if !m[pinfo.Type] {
				log.Warnf("Unhandled type %s: %s", pinfo.Ticker, pinfo.Type)
			}
		}
		p.positions[op.Figi] = pinfo
	}

	if op.IsTrading() {
		deal = &schema.Deal{
			Date:       op.DateParsed,
			Price:      schema.NewCValue(op.Price, op.Currency),
			Commission: op.Commission.Value,
		}

		// bug or feature?
		// op.Quantity reflects the whole order size;
		// if the order is only partially completed, sum(op.Trades.Quantity) < op.Quantity
		for _, trade := range op.Trades {
			deal.Quantity += int(trade.Quantity)
		}
		if op.OperationType == "Sell" {
			deal.Quantity = -deal.Quantity
		}

		// op.Payment is negative for Buy
		// deal.Quantity is positive for Buy
		// deal.Price is always positive
		// Commission is not included in Payment
		deal.Accrued = -op.Payment - deal.Price.Value*float64(deal.Quantity)

	} else if op.OperationType == "BrokerCommission" {
		// negative
		pinfo.AccumulatedIncome.Value += op.Payment

	} else if op.IsPayment() {
		// income - positive, taxes - negative
		pinfo.AccumulatedIncome.Value += op.Payment
		pinfo.Dividends = append(pinfo.Dividends,
			&schema.Dividend{
				Date:  op.DateParsed,
				Value: op.Payment,
			})
	} else if op.OperationType == "Tax" {
		// negative
		pinfo.AccumulatedIncome.Value += op.Payment
	} else {
		log.Warnf("Unprocessed transaction %v", op)
	}

	return
}

// =============================================================================

/*

 Balance consists of:

 USD
 + Assets:
    1. Cash balance
        1.1 Direct payins
        1.2 Exchanges
        1.3 Sold stocks
        1.4 - Bought stocks
        1.5 - Service commissions
        1.6 - Tax
        1.7 Dividends & coupons
    2. Open USD positions
 - Payins
    3. Directs payins
    4. Exchanges

RUB
 + Assets:
    1. Cash balance
        1.1 Direct payins
        1.3 Sold stocks & dollars
        1.4 - Bought stocks & dollars
        1.5 - Service commissions
        1.6 - Tax
        1.7 Dividends & coupons
    2. Open RUB positions
 - Payins:
    3. Direct payins
    5. - Exchanged money

*/

func (p *Portfolio) addOpToBalance(bal *schema.Balance, op schema.Operation) {
	if op.IsTrading() || op.OperationType == "BrokerCommission" {
		// not accounted here

	} else if op.IsPayment() {
		// 1.7
		bal.Assets[op.Currency].Value += op.Payment
	} else if op.OperationType == "PayIn" {
		// 1.1
		bal.Assets[op.Currency].Value += op.Payment
		// 3
		bal.Payins[op.Currency].Value += op.Payment
	} else if op.OperationType == "ServiceCommission" {
		bal.Commissions[op.Currency].Value += op.Payment
		// 1.5
		bal.Assets[op.Currency].Value -= -op.Payment
	} else if op.OperationType == "Tax" {
		// 1.6
		bal.Assets[op.Currency].Value -= -op.Payment
	} else {
		log.Warnf("Unprocessed transaction 2 %v", op)
	}
}

func (p *Portfolio) addDealToBalance(bal *schema.Balance, figi string, deal *schema.Deal) {
	pinfo := p.positions[figi]
	if pinfo.Figi == schema.FigiUSD {
		// Exchanges
		// 1.2
		bal.Assets["USD"].Value += float64(deal.Quantity)
		// 4
		bal.Payins["USD"].Value += float64(deal.Quantity)
		// 5
		bal.Payins["RUB"].Value -= deal.Value()
	}
	// 1.3, 1.4, 2
	bal.Assets[deal.Price.Currency].Value -= deal.Value() - deal.Commission
}

// =============================================================================

func (p *Portfolio) getOpenPortion(pinfo *schema.PositionInfo) *schema.Portion {
	if len(pinfo.Portions) == 0 {
		return nil
	}

	po := pinfo.Portions[len(pinfo.Portions)-1]
	if po.IsClosed {
		return nil
	}

	return po
}

func (p *Portfolio) addToPortions(pinfo *schema.PositionInfo, deal *schema.Deal) {
	po := p.getOpenPortion(pinfo)
	if po == nil {
		po = &schema.Portion{
			Balance: schema.NewCValue(0, deal.Price.Currency),
			AvgDate: deal.Date,
		}
		pinfo.Portions = append(pinfo.Portions, po)
	}

	pinfo.OpenQuantity += deal.Quantity
	pinfo.OpenSpent += deal.Value()

	if deal.Quantity > 0 { // buy
		po.CheckNoSplitSells(pinfo.Ticker)

		// TODO think again is this correct?
		mult := deal.Value() / pinfo.OpenSpent

		biasDays := int(math.Round(deal.Date.Sub(po.AvgDate).Hours() * mult / 24))
		po.AvgDate.AddDate(0, 0, biasDays)

		po.AvgPrice.Value = deal.Price.Value*mult + po.AvgPrice.Value*(1-mult)
		po.Buys = append(po.Buys, deal)

	} else { // sell

		if pinfo.OpenQuantity < 0 {
			log.Fatalf("negative balance? %v", pinfo)
		}

		if pinfo.OpenQuantity > 0 {
			// How to better handle it?
			// 1. split sells. try and merge. those wont split between days,
			//    so wont cross portion period boundaries
			//
			// 2. true partial sells. options?
			//    2.1 sell all, buy some back
			//    2.2 ?
			//
			po.SplitSells = append(po.SplitSells, deal)

		} else {
			if len(po.SplitSells) > 0 {
				// new "superdeal"
				sdeal := *deal
				sval := sdeal.Value()

				for _, psell := range po.SplitSells {
					sdeal.Quantity += psell.Quantity
					sdeal.Accrued += psell.Accrued
					sdeal.Commission += psell.Commission
					sval += psell.Value()
				}

				sdeal.Price.Value = (sval - sdeal.Accrued) / float64(sdeal.Quantity)
				deal = &sdeal
			}

			pinfo.OpenSpent = 0

			// complete sell
			po.Close = deal
			po.IsClosed = true
		}
	}
}

func (p *Portfolio) getAccrued(pinfo *schema.PositionInfo, date time.Time) float64 {
	// Accrued value cannot be fetched for date != Now
	if pinfo.Type != schema.InsTypeBond {
		return 0
	}
	if !p.config.enableAccrued {
		return 0
	}
	if time.Now().Sub(date).Hours() > 24 {
		return 0
	}

	accrued, ok := p.accrued[pinfo.Figi]
	if !ok {
		log.Warnf("missing accrued value for %s, balance is inaccurate", pinfo.Figi)
	}
	return accrued
}

func (p *Portfolio) makeOpenDeal(pinfo *schema.PositionInfo, date time.Time, pricef schema.Pricef0, setClose bool) *schema.Deal {
	po := p.getOpenPortion(pinfo)
	if po == nil {
		return nil
	}

	po.CheckNoSplitSells(pinfo.Ticker)

	deal := &schema.Deal{
		Date:     date,
		Price:    schema.NewCValue(pricef()+p.getAccrued(pinfo, date), po.Balance.Currency),
		Quantity: -pinfo.OpenQuantity,
	}

	if setClose {
		po.Close = deal
		pinfo.OpenDeal = deal
	}

	return deal
}

func (p *Portfolio) getYield(figi string, t1, t2 time.Time) float64 {
	p1 := p.cc.Get(figi, t1)
	p2 := p.cc.Get(figi, t2)
	return p2/p1*100 - 100
}

func (p *Portfolio) getBenchmark(ticker, typ, currency string) string {
	if typ == schema.InsTypeBond {
		if currency == "RUB" {
			return "FXRB"
		}
		// eurobond.. dont even know
		return ""
	}

	if typ == schema.InsTypeStock {
		if currency == "RUB" {
			return "FXRL"
		}

		// sorry thats all I personally had so far ;)
		fxitTickers := map[string]bool{
			"MSFT": true,
			"NVDA": true,
		}

		if fxitTickers[ticker] {
			return "FXIT"
		}

		return "FXUS"
	}

	return ""
}

func (p *Portfolio) makePortionYields(pinfo *schema.PositionInfo) {
	for _, po := range pinfo.Portions {
		var expense float64

		profit := schema.NewCValue(-po.Close.Value(), po.Close.Price.Currency)

		for _, div := range pinfo.Dividends {
			if div.Date.Before(po.Buys[0].Date) {
				continue
			}
			if div.Date.After(po.Close.Date) {
				// TODO not quite right. Dividends come with delay
				continue
			}
			profit.Value += div.Value
		}

		for _, deal := range po.Buys {
			expense += deal.Value()
			expense += -deal.Commission
		}
		expense += -po.Close.Commission

		po.Yield = profit.Div(expense / 100)
		po.Yield.Value -= 100

		po.YieldAnnual = po.Yield.Value * 365 / (po.Close.Date.Sub(po.AvgDate).Hours() / 24)

		po.Balance = profit
		po.Balance.Value -= expense

		// now compare with the market ETF
		bench := p.getBenchmark(pinfo.Ticker, pinfo.Type, po.Balance.Currency)
		if bench != "" {
			po.YieldMarket = p.getYield(p.getFigi(bench), po.AvgDate, po.Close.Date)
		}
	}
}

// =============================================================================

func (p *Portfolio) processOperations(cb func(*schema.Balance, time.Time) bool) *schema.Balance {
	p.preprocessOperations(beginning)

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

		deal := p.processOperation(op)
		if deal != nil {
			pinfo := p.positions[op.Figi]
			pinfo.Deals = append(pinfo.Deals, deal)

			p.addToPortions(pinfo, deal)
			p.addDealToBalance(bal, pinfo.Figi, deal)
		}

		p.addOpToBalance(bal, op)

		log.Debugf("new balance: %s", bal)
	}

	return bal
}

func (p *Portfolio) openDealsBalancePerType(time time.Time, pricef1 schema.Pricef1) map[string]*schema.Balance {
	m := make(map[string]*schema.Balance)
	total := schema.NewBalance()
	m[""] = total

	for _, pinfo := range p.positions {
		pricef0 := schema.PriceCurry0(pricef1, pinfo.Figi)

		od := p.makeOpenDeal(pinfo, time, pricef0, true)

		log.Debugf("open deal %s %s %s", pinfo.Figi, pinfo.Ticker, od)

		if od == nil || pinfo.Figi == schema.FigiUSD {
			continue
		}

		bal := m[pinfo.Type]
		if bal == nil {
			bal = schema.NewBalance()
			m[pinfo.Type] = bal
		}

		p.addDealToBalance(bal, pinfo.Figi, od)
		p.addDealToBalance(total, pinfo.Figi, od)
	}

	return m
}

func (p *Portfolio) openDealsBalance(time time.Time, pricef1 schema.Pricef1) *schema.Balance {
	m := p.openDealsBalancePerType(time, pricef1)
	log.Debugf("current asset balance: %s", m[""])
	return m[""]
}

func (p *Portfolio) Collect(at time.Time) {
	var once sync.Once

	p.config.enableAccrued = true

	p.processPortfolio()

	p.cash = p.processOperations(func(bal *schema.Balance, opTime time.Time) bool {
		once.Do(func() {
			p.cc = candles.NewCandleCache(p.client, opTime, "week")
		})
		return opTime.Before(at)
	})

	m := p.openDealsBalancePerType(at, p.cc.Pricef1(at))

	p.funds = m[schema.InsTypeEtf]
	p.bonds = m[schema.InsTypeBond]
	p.stocks = m[schema.InsTypeStock]
	p.totals = m[""]
	p.totals.Add(*p.cash)

	for _, pinfo := range p.positions {
		p.makePortionYields(pinfo)
	}
}

// =============================================================================

func (p *Portfolio) ListDeals(start time.Time) {
	empty := true

	p.processPortfolio()

	p.preprocessOperations(start)

	deals := schema.NewBalance()
	comms := schema.NewBalance()
	for _, op := range p.data.ops {
		if op.Status != "Done" {
			// cancelled declined etc
			// noone is interested in that
			continue
		}

		if op.Figi != "" {
			op.Ticker = p.getTicker(op.Figi)
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
	_, dealsT, _ := deals.GetTotal(usdrate, 0)
	_, commsT, _ := comms.GetTotal(usdrate, 0)
	fmt.Printf("   percentage: %.2f%%\n", commsT/dealsT*100)
}

// =============================================================================

func (p *Portfolio) summarize( /* const */ bal schema.Balance, t time.Time, format string) {
	pricef1 := p.cc.Pricef1(t)
	obal := p.openDealsBalance(t, pricef1)
	obal.Add(bal)
	fmt.Print(obal.ToString(t, pricef1(schema.FigiUSD), 0, format))
}

func (p *Portfolio) ListBalances(start time.Time, period, format string) {
	p.processPortfolio()

	p.cc = candles.NewCandleCache(p.client, start, period)

	// just for time reference, can be any figi
	candles := p.cc.List(schema.FigiUSD).Payload.Candles

	cidx := 0
	num := len(candles)

	if num == 0 {
		log.Debug("No data for this period")
		return
	}

	bal := p.processOperations(func(bal *schema.Balance, opTime time.Time) bool {

		// process all candles before opTime

		for ; cidx < num; cidx += 1 {
			nextTime := candles[cidx].TimeParsed
			if opTime.Before(nextTime) {
				break
			}
			p.summarize(*bal, nextTime, format)
		}

		return true
	})

	// process all candles after the last operation

	for ; cidx < num; cidx += 1 {
		nextTime := candles[cidx].TimeParsed
		p.summarize(*bal, nextTime, format)
	}

	// current balance

	p.summarize(*bal, time.Now(), format)
}
