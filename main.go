package main

import (
	"flag"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aladh/domain_name_checker/rdap"
)

func init() {
	flag.Parse()
}

func main() {
	domains := strings.Split(flag.Arg(0), ",")

	err := rdap.Initialize()
	if err != nil {
		log.Fatalf("error initialzing rdap: %s", err)
	}

	tldRegex := regexp.MustCompile("^.*\\.(.*)$")
	tldsChan := make(chan *tld)

	go partitionBy(domains, tldsChan, func(domain string) string {
		matches := tldRegex.FindStringSubmatch(domain)
		return matches[1]
	})

	availableChan := checkAvailability(tldsChan)
	for domain := range availableChan {
		log.Println(domain)
	}
}

func checkAvailability(tldsChan chan *tld) chan string {
	availableChan := make(chan string, 1)

	go checkAvailabilityAsync(tldsChan, availableChan)

	return availableChan
}

func checkAvailabilityAsync(tldsChan chan *tld, availableChan chan string) {
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

func checkAvailabilityForTld(t *tld, availableChan chan string) {
	log.Printf("starting worker for tld: %s\n", t.Name)

	for domain := range t.Domains {
		log.Printf("processing domain: %s\n", domain)
		expiryDate, err := rdap.ExpiryDate(domain)
		if err != nil {
			log.Printf("error checking expiry date for domain %s: %s\n", domain, err)
			continue
		}

		log.Printf("domain %s expires on %s\n", domain, expiryDate)
		time.Sleep(200 * time.Millisecond)
	}
}

type tld struct {
	Name    string
	Domains chan string
}

func NewTld(name string) *tld {
	const tldBuffer = 100

	return &tld{
		Name:    name,
		Domains: make(chan string, tldBuffer),
	}
}

func partitionBy(domainsChan []string, tldsChan chan *tld, partitionKey func(string) string) {
	tldMap := make(map[string]*tld)

	for _, domain := range domainsChan {
		tldName := partitionKey(domain)

		if tldForDomain, ok := tldMap[tldName]; ok {
			tldForDomain.Domains <- domain
		} else {
			tldForDomain = NewTld(tldName)
			tldMap[tldName] = tldForDomain
			tldsChan <- tldForDomain
			tldForDomain.Domains <- domain
		}
	}

	for _, v := range tldMap {
		close(v.Domains)
	}
	close(tldsChan)
}
