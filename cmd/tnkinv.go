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
	token, period string

	start, at time.Time
}

func parseCmdline() (string, config) {
	if len(os.Args) < 2 {
		usage()
		log.Fatal("no cmd provided")
	}

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

	period := fs.String("period", "month", "story period")
	start := fs.String("start", "", "starting point in time (format: 1922/12/28; default: year ago)")
	atTime := fs.String("at", "", "point in time (default: now). Not supported yet")

	fs.Parse(os.Args[2:])

	startParsed := time.Now().AddDate(-1, 0, 0)
	if *start != "" {
		var err error
		startParsed, err = time.Parse("2006/01/02", *start)
		if err != nil {
			usage()
			log.Fatalf("unrecognized date %s", *start)
		}
	}

	atTimeParsed := time.Now()
	if *atTime != "" {
		var err error
		atTimeParsed, err = time.Parse("2006/01/02", *atTime)
		if err != nil {
			usage()
			log.Fatalf("unrecognized date %s", *atTime)
		}
	}

	return cmd, config{
		token:  *token,
		period: *period,
		start:  startParsed,
		at:     atTimeParsed,
	}
}

func usage() {
	fmt.Printf("usage:\n" +
		"\t tnkinv {show|story|deals} --token file_with_token [--period week|month|year]\n")
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
		port.Collect(cfg.at)
		port.Print()
		return
	}

	if cmd == "deals" {
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
		port.ListBalances(cfg.start, cfg.period)
		return
	}
}
