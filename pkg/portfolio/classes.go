package portfolio

import (
	log "github.com/sirupsen/logrus"

	"../schema"
)

func getInstrumentType(typ string, ticker string) schema.InsType {
	m := map[schema.InsType]bool{
		schema.InsTypeEtf:      true,
		schema.InsTypeStock:    true,
		schema.InsTypeBond:     true,
		schema.InsTypeCurrency: true,
	}
	if !m[schema.InsType(typ)] {
		log.Warnf("Unhandled type %s: %s", ticker, typ)
	}
	return schema.InsType(typ)
}

func getBenchmark(ticker string, typ schema.InsType, currency string) string {
	m := map[string]string{
		schema.InsTypeBond + "RUB":  "VTBB",
		schema.InsTypeBond + "USD":  "FXRU",
		schema.InsTypeStock + "RUB": "FXRL",
		schema.InsTypeStock + "USD": "FXUS", // see below
	}

	if bench, ok := m[string(typ)+currency]; ok {
		if bench == "FXUS" {
			// sorry thats all I personally had so far ;)
			fxitTickers := map[string]bool{
				"MSFT": true,
				"NVDA": true,
			}

			if fxitTickers[ticker] {
				return "FXIT"
			}
		}
		return bench
	}

	if typ == schema.InsTypeEtf {
		// ETF is benchmark itself
	}

	return ""
}

func getSection(ticker string, typ schema.InsType, currency string) schema.Section {
	m := map[string]schema.Section{
		schema.InsTypeBond + "RUB": schema.BondRub,
		schema.InsTypeBond + "USD": schema.BondUsd,

		schema.InsTypeStock + "RUB": schema.StockRub,
		schema.InsTypeStock + "USD": schema.StockUsd,

		schema.InsTypeCurrency + "RUB": schema.CashRub,
		schema.InsTypeCurrency + "USD": schema.CashUsd,
	}
	if s, ok := m[string(typ)+currency]; ok {
		return s
	}

	if typ == schema.InsTypeEtf {
		mt := map[string]schema.Section{
			"VTBB": schema.BondRub,
			"FXRB": schema.BondRub,

			"FXRU": schema.BondUsd,

			"SBMX": schema.StockRub,
			"FXRL": schema.StockRub,

			"AKNX": schema.StockUsd,
			"FXIT": schema.StockUsd,
			"FXUS": schema.StockUsd,

			"FXMM": schema.CashRub,
			"FXTB": schema.CashUsd,
		}
		if s, ok := mt[ticker]; ok {
			return s
		}
		log.Warnf("Uncatched ETF %s", ticker)
	}

	return ""
}
