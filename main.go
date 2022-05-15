package main

import (
	"flag"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"
)

func init() {
	flag.Parse()
}

func main() {
	domains := strings.Split(flag.Arg(0), ",")

	domainsChan := make(chan string)
	go func() {
		for _, domain := range domains {
			domainsChan <- domain
		}

		close(domainsChan)
	}()

	tldRegex := regexp.MustCompile("^.*\\.(.*)$")
	tldsChan := make(chan *tld)
	go partitionBy(domainsChan, tldsChan, func(domain string) string {
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

		go func(t1 *tld) {
			defer wg.Done()

			checkAvailabilityForTld(t1, availableChan)
		}(t)
	}

	wg.Wait()
	close(availableChan)
}

func checkAvailabilityForTld(t *tld, availableChan chan string) {
	log.Printf("starting worker for tld: %s\n", t.Name)

	for domain := range t.Domains {
		log.Printf("processing domain: %s\n", domain)
		time.Sleep(1 * time.Second)
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

func partitionBy(domainsChan chan string, tldsChan chan *tld, partitionKey func(string) string) {
	tldMap := make(map[string]*tld)

	for domain := range domainsChan {
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
