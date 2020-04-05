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
