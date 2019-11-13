package main

import (
	"flag"
	"log"

	"../pkg/client"
	"../pkg/portfolio"
)

type config struct {
	token     string
	sandToken string
}

func parseCmdline() config {
	token := flag.String("token", "", "API token")
	sandToken := flag.String("sandtoken", "", "sandbox API token")

	flag.Parse()

	return config{
		token:     *token,
		sandToken: *sandToken,
	}
}

func main() {
	cfg := parseCmdline()

	if cfg.sandToken != "" {
		c := client.NewClient(cfg.sandToken)
		c.TrySandbox()
		c.Stop()
	}

	if cfg.token == "" {
		log.Printf("no token provided")
		return
	}

	port := portfolio.NewPortfolio(client.NewClient(cfg.token))
	port.Collect()
	port.Print()
}
