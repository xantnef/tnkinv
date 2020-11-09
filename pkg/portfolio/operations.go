package portfolio

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"

	"../schema"
)

func readOperations(fname string) (ops []schema.Operation) {
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Fatal(err)
	}

	err = json.Unmarshal(data, &ops)
	if err != nil {
		log.Fatal(err)
	}

	return ops
}

func (p *Portfolio) getOperations(start time.Time) (ops []schema.Operation) {
	if p.config.fictFile == "" {
		for _, acc := range p.accs {
			resp := p.client.RequestOperations(start, acc)
			ops = append(ops, resp.Payload.Operations...)
		}
	}

	if p.config.opsFile != "" {
		ops = append(ops, readOperations(p.config.opsFile)...)
	}

	if p.config.fictFile != "" {
		ops = append(ops, fetchFictives(p.client, p.cc, p.config.fictFile)...)
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
	return
}
