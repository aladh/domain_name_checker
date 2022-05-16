package main

import (
	"flag"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/aladh/domain_name_checker/domain"
	"github.com/aladh/domain_name_checker/rdap"
)

func init() {
	flag.Parse()
}

func main() {
	domains := *new([]*domain.Domain)
	for _, name := range strings.Split(flag.Arg(0), ",") {
		domains = append(domains, domain.New(name))
	}

	err := rdap.Initialize()
	if err != nil {
		log.Fatalf("error initialzing rdap: %s", err)
	}

	tldsChan := make(chan *tld)

	go partitionBy(domains, tldsChan, func(domain domain.Domain) string {
		return domain.Tld
	})

	availableChan := checkAvailability(tldsChan)
	for d := range availableChan {
		log.Println(d.Name)
	}
}

func checkAvailability(tldsChan chan *tld) chan *domain.Domain {
	availableChan := make(chan *domain.Domain)

	go checkAvailabilityAsync(tldsChan, availableChan)

	return availableChan
}

func checkAvailabilityAsync(tldsChan chan *tld, availableChan chan *domain.Domain) {
	wg := sync.WaitGroup{}

	for t := range tldsChan {
		wg.Add(1)

		go func(t *tld) {
			defer wg.Done()

			checkAvailabilityForTld(t, availableChan)
		}(t)
	}

	wg.Wait()
	close(availableChan)
}

func checkAvailabilityForTld(t *tld, availableChan chan *domain.Domain) {
	log.Printf("starting worker for tld: %s\n", t.Name)

	for d := range t.Domains {
		log.Printf("processing domain: %s\n", d.Name)
		expiryDate, err := rdap.ExpiryDate(d)
		if err != nil {
			log.Printf("error checking expiry date for domain %s: %s\n", d.Name, err)
			continue
		}

		log.Printf("domain %s expires on %s\n", d.Name, expiryDate)
		time.Sleep(200 * time.Millisecond)
	}
}

type tld struct {
	Name    string
	Domains chan *domain.Domain
}

func NewTld(name string) *tld {
	const tldBuffer = 100

	return &tld{
		Name:    name,
		Domains: make(chan *domain.Domain, tldBuffer),
	}
}

func partitionBy(domainsChan []*domain.Domain, tldsChan chan *tld, partitionKey func(domain.Domain) string) {
	tldMap := make(map[string]*tld)

	for _, d := range domainsChan {
		tldName := partitionKey(*d)

		if tldForDomain, ok := tldMap[tldName]; ok {
			tldForDomain.Domains <- d
		} else {
			tldForDomain = NewTld(tldName)
			tldMap[tldName] = tldForDomain
			tldsChan <- tldForDomain
			tldForDomain.Domains <- d
		}
	}

	for _, v := range tldMap {
		close(v.Domains)
	}
	close(tldsChan)
}
