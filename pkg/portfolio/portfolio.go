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

	instruments map[string]schema.Instrument // key=figi
	positions   map[string]*schema.PositionInfo

	accrued map[string]float64

	figisSorted []string

	cash, funds, bonds, stocks, totals *schema.Balance // may be nil!

	config struct {
		enableAccrued bool
	}
}

func NewPortfolio(c *client.MyClient, accs []string) *Portfolio {
	return &Portfolio{
		client: c,
		accs:   accs,

		instruments: make(map[string]schema.Instrument),
		positions:   make(map[string]*schema.PositionInfo),
		accrued:     make(map[string]float64),
	}
}

// =============================================================================

func (p *Portfolio) insByFigi(figi string) schema.Instrument {
	ins, ok := p.instruments[figi]
	if !ok {
		ins = p.client.RequestByFigi(figi)
		p.instruments[figi] = ins
	}
	log.Debug(ins)
	return ins
}

func (p *Portfolio) insByTicker(ticker string) schema.Instrument {
	for _, ins := range p.instruments {
		if ins.Ticker == ticker {
			return ins
		}
	}

	ins := p.client.RequestByTicker(ticker)
	p.instruments[ins.Figi] = ins
	return ins
}

// =============================================================================

func (p *Portfolio) processPortfolio() {
	for _, acc := range p.accs {
		pfResp := p.client.RequestPortfolio(acc)
		for _, pos := range pfResp.Payload.Positions {
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

func repaymentMultiplier(pinfo *schema.PositionInfo, t time.Time) float64 {
	idx := sort.Search(len(pinfo.Repayments), func(i int) bool {
		return pinfo.Repayments[i].Time.After(t)
	})

	if idx < len(pinfo.Repayments) {
		return pinfo.Repayments[idx].Mult
	}

	return 1
}

func (p *Portfolio) addRepayment(figi string, t time.Time, value float64) {
	pinfo := p.positions[figi]
	pinfo.Repayments = append(pinfo.Repayments,
		&schema.RepaymentPoint{
			Time: t,
		})

	for _, rep := range pinfo.Repayments {
		rep.Mult += value
	}
}

func (p *Portfolio) addStaticRepayments() {
	if p.positions["BBG00GW0RM55"] != nil {
		p.addRepayment("BBG00GW0RM55",
			time.Date(2019, 12, 10, 7, 0, 0, 0, time.UTC),
			83)
		p.addRepayment("BBG00GW0RM55",
			time.Date(2020, 3, 10, 7, 0, 0, 0, time.UTC),
			83)
	}
}

func (p *Portfolio) calculateRepayments() {
	amounts := make(map[string]int)

	for _, op := range p.data.ops {
		if op.Status != "Done" {
			continue
		}

		if op.IsTrading() {
			amounts[op.Figi] += operationQuantity(op)
			continue
		}

		if op.OperationType == "PartRepayment" {
			p.addRepayment(op.Figi, op.DateParsed, op.Payment/float64(amounts[op.Figi]))
			continue
		}
	}

	// Temporary fixup:
	// The problem is that the solution doesn't work for
	//  - amortized bonds
	//  - that were open at some t1 (point we want to know balance at)
	//  - but were all sold at some point t2
	//  - and there were more repayments from t2 till now
	// ..because there seems to be no way to get their partrepayment stats after the selling point
	// Maybe extrapolate the previous repayments?
	p.addStaticRepayments()

	// Now normalize the multipliers
	for _, pinfo := range p.positions {
		for _, rep := range pinfo.Repayments {
			log.Debugf("repayment %s at %s: %f/%d", pinfo.Ins.Name, rep.Time, rep.Mult, pinfo.Ins.FaceValue)
			rep.Mult = (rep.Mult + float64(pinfo.Ins.FaceValue)) / float64(pinfo.Ins.FaceValue)
		}
	}
}

func (p *Portfolio) addPosition(op schema.Operation) *schema.PositionInfo {
	if pinfo, exists := p.positions[op.Figi]; exists {
		return pinfo
	}

	pinfo := &schema.PositionInfo{
		Ins: p.insByFigi(op.Figi),

		AccumulatedIncome: schema.NewCValue(0, op.Currency),
	}
	pinfo.Ins.Type = getInstrumentType(op.InstrumentType, pinfo.Ins.Ticker)

	p.positions[op.Figi] = pinfo
	return pinfo
}

func operationQuantity(op schema.Operation) int {
	quantity := 0
	// bug or feature?
	// op.Quantity reflects the whole order size;
	// if the order is only partially completed, sum(op.Trades.Quantity) < op.Quantity
	for _, trade := range op.Trades {
		quantity += int(trade.Quantity)
	}
	if op.OperationType == "Sell" {
		quantity = -quantity
	}
	return quantity
}

func (p *Portfolio) accountOperation(pinfo *schema.PositionInfo, op schema.Operation) (deal *schema.Deal) {
	log.Debugf("%s", op)

	if op.IsTrading() {
		deal = &schema.Deal{
			Date:       op.DateParsed,
			Price:      schema.NewCValue(op.Price, op.Currency),
			Quantity:   operationQuantity(op),
			Commission: op.Commission.Value,
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

func xchgrate(cc *candles.CandleCache, currency string, t time.Time) float64 {
	if currency == "RUB" {
		return 1
	}
	if currency == "USD" {
		return cc.GetOnDay(schema.FigiUSD, t)
	}

	return 0
}

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
        1.7 Dividends, coupons & repayments
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
        1.7 Dividends, coupons & repayments
    2. Open RUB positions
 - Payins:
    3. Direct payins
    5. - Exchanged money

*/

func addOpToBalance(bal *schema.Balance, op schema.Operation, cc *candles.CandleCache) {
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

		// add total payin
		bal.Payins["all"].Value += op.Payment * xchgrate(cc, op.Currency, op.DateParsed)

	} else if op.OperationType == "ServiceCommission" {

		bal.Commissions[op.Currency].Value += op.Payment
		// add total
		bal.Commissions["all"].Value += op.Payment * xchgrate(cc, op.Currency, op.DateParsed)

		// 1.5
		bal.Assets[op.Currency].Value -= -op.Payment

	} else if op.OperationType == "Tax" {
		// 1.6
		bal.Assets[op.Currency].Value -= -op.Payment
	} else {
		log.Warnf("Unprocessed transaction 2 %v", op)
	}
}

func addDealToBalance(bal *schema.Balance, figi string, deal *schema.Deal) {
	if figi == schema.FigiUSD {
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
		po.CheckNoSplitSells(pinfo.Ins.Ticker)

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
	if pinfo.Ins.Type != schema.InsTypeBond {
		return 0
	}
	if !p.config.enableAccrued {
		return 0
	}
	if time.Now().Sub(date).Hours() > 24 {
		return 0
	}

	accrued, ok := p.accrued[pinfo.Ins.Figi]
	if !ok {
		log.Warnf("missing accrued value for %s, balance is inaccurate", pinfo.Ins.Figi)
	}
	return accrued
}

func (p *Portfolio) getPrice(pinfo *schema.PositionInfo, t time.Time) float64 {
	return p.cc.Get(pinfo.Ins.Figi, t)*repaymentMultiplier(pinfo, t) + p.getAccrued(pinfo, t)
}

func (p *Portfolio) makeOpenDeal(pinfo *schema.PositionInfo, date time.Time, setClose bool) *schema.Deal {
	po := p.getOpenPortion(pinfo)
	if po == nil {
		return nil
	}

	po.CheckNoSplitSells(pinfo.Ins.Ticker)

	deal := &schema.Deal{
		Date:     date,
		Price:    schema.NewCValue(p.getPrice(pinfo, date), po.Balance.Currency),
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
		bench := getBenchmark(pinfo.Ins.Ticker, pinfo.Ins.Type, po.Balance.Currency)
		if bench != "" {
			po.YieldMarket = p.getYield(p.insByTicker(bench).Figi, po.AvgDate, po.Close.Date)
		}
	}
}

// =============================================================================

func (p *Portfolio) processOperations(cb func(*schema.Balance, time.Time) bool) *schema.Balance {
	p.preprocessOperations(beginning)

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

		if op.Figi != "" {
			pinfo := p.positions[op.Figi]
			deal := p.accountOperation(pinfo, op)
			if deal != nil {
				pinfo.Deals = append(pinfo.Deals, deal)
				p.addToPortions(pinfo, deal)
				addDealToBalance(bal, pinfo.Ins.Figi, deal)
			}
		}

		addOpToBalance(bal, op, p.cc)

		log.Debugf("new balance: %s", bal)
	}

	for _, pinfo := range p.positions {
		if len(pinfo.Deals) == 0 {
			delete(p.positions, pinfo.Ins.Figi)
		}
	}

	return bal
}

func (p *Portfolio) openDealsBalancePerType(time time.Time) map[schema.InsType]*schema.Balance {
	m := make(map[schema.InsType]*schema.Balance)
	total := schema.NewBalance()
	m[""] = total

	for _, pinfo := range p.positions {
		od := p.makeOpenDeal(pinfo, time, true)

		log.Debugf("open deal %s %s %s", pinfo.Ins.Figi, pinfo.Ins.Ticker, od)

		if od == nil || pinfo.Ins.Figi == schema.FigiUSD {
			continue
		}

		bal := m[pinfo.Ins.Type]
		if bal == nil {
			bal = schema.NewBalance()
			m[pinfo.Ins.Type] = bal
		}

		addDealToBalance(bal, pinfo.Ins.Figi, od)
		addDealToBalance(total, pinfo.Ins.Figi, od)
	}

	return m
}

func (p *Portfolio) openDealsBalance(time time.Time) *schema.Balance {
	m := p.openDealsBalancePerType(time)
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

	m := p.openDealsBalancePerType(at)

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

func (p *Portfolio) summarize( /* const */ bal schema.Balance, t time.Time, format string) {
	obal := p.openDealsBalance(t)
	obal.Add(bal)
	obal.CalcAllAssets(p.cc.Get(schema.FigiUSD, t), 0)
	fmt.Print(obal.ToString(t.Format("2006/01/02"), format))
}

func (p *Portfolio) ListBalances(start time.Time, period, format string) {
	p.processPortfolio()

	p.cc = candles.NewCandleCache(p.client, start, period)

	candleTimes := p.cc.ListTimes()

	cidx := 0
	num := len(candleTimes)

	if num == 0 {
		log.Debug("No data for this period")
		return
	}

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

	// process all candles after the last operation

	for ; cidx < num; cidx += 1 {
		nextTime := candleTimes[cidx]
		p.summarize(*bal, nextTime, format)
	}

	p.summarize(*bal, time.Now().Add(24*time.Hour), format)
}
