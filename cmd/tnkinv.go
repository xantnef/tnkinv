package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"../pkg/client"
	"../pkg/portfolio"
)

type config struct {
	token, period, format string

	start, at time.Time
}

func parseCmdline() (string, config) {
	if len(os.Args) < 2 {
		usage()
		log.Fatal("no cmd provided")
	}

	cfg := config{}

	cmd := os.Args[1]
	cmds := map[string]bool{
		"sandbox": true,
		"show":    true,
		"story":   true,
		"deals":   true,
	}

	if !cmds[cmd] {
		usage()
		log.Fatalf("unknown command %s", cmd)
	}

	fs := flag.NewFlagSet("", flag.ExitOnError)
	token := fs.String("token", "", "API token")
	format := fs.String("format", "human", "output format")

	period := fs.String("period", "month", "story period")
	start := fs.String("start", "", "starting point in time (format: 1922/12/28; default: year ago)")
	atTime := fs.String("at", "", "point in time (default: now). Not supported yet")

	fs.Parse(os.Args[2:])

	cfg.token = *token
	cfg.period = *period
	cfg.format = *format

	if *start != "" {
		var err error
		cfg.start, err = time.Parse("2006/01/02", *start)
		if err != nil {
			usage()
			log.Fatalf("unrecognized date %s", *start)
		}
	}

	if *atTime != "" {
		var err error
		cfg.at, err = time.Parse("2006/01/02", *atTime)
		if err != nil {
			usage()
			log.Fatalf("unrecognized date %s", *atTime)
		}
	}

	return cmd, cfg
}

func usage() {
	fmt.Printf("usage:\n" +
		"\t tnkinv {subcmd} [params] --token file_with_token \n" +
		"\t   common params: \n" +
		"\t     --format human|table \n" +
		"\t   subcmds: \n" +
		"\t     show   [--at 1922/12/28 (default: today)] \n" +
		"\t     story  [--start 1901/01/01 (default: year ago)] \n" +
		"\t            [--period day|week|month (default: month)] \n" +
		"\t     deals  [--start 1901/01/01 (default: none)] \n" +
		"\t            [--period day|week|month|all (default: month)] \n" +
		"\t     sandbox \n")
}

func main() {
	cmd, cfg := parseCmdline()

	if cfg.token == "" {
		usage()
		log.Fatal("no token provided")
	}

	if cmd == "sandbox" {
		c := client.NewClient(cfg.token)
		c.TrySandbox()
		c.Stop()
		return
	}

	port := portfolio.NewPortfolio(client.NewClient(cfg.token))

	if cmd == "show" {
		if cfg.at.IsZero() {
			cfg.at = time.Now()
		}
		port.Collect(cfg.at)
		port.Print()
		return
	}

	if cmd == "deals" {
		if !cfg.start.IsZero() {
			port.ListDeals(cfg.start)
			return
		}

		since := time.Now()

		if cfg.period == "day" {
			since = since.AddDate(0, 0, -1)
		}
		if cfg.period == "week" {
			since = since.AddDate(0, 0, -7)
		}
		if cfg.period == "month" {
			since = since.AddDate(0, -1, 0)
		}
		if cfg.period == "all" {
			since = time.Time{}
		}

		port.ListDeals(since)
		return
	}

	if cmd == "story" {
		if cfg.start.IsZero() {
			cfg.start = time.Now().AddDate(-1, 0, 0)
		}
		port.ListBalances(cfg.start, cfg.period, cfg.format)
		return
	}
}
