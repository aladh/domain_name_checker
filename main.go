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
	tldChanChan := make(chan chan string)
	go partitionBy(domainsChan, tldChanChan, func(domain string) string {
		matches := tldRegex.FindStringSubmatch(domain)
		return matches[1]
	})

	availableChan := checkAvailability(tldChanChan)

	for domain := range availableChan {
		log.Println(domain)
	}
}

func checkAvailability(tldChanChan chan chan string) chan string {
	availableChan := make(chan string, 1)

	go checkAvailabilityAsync(tldChanChan, availableChan)

	return availableChan
}

func checkAvailabilityAsync(tldChanChan chan chan string, availableChan chan string) {
	wg := sync.WaitGroup{}

	for tldChan := range tldChanChan {
		wg.Add(1)

		go func(tldChan chan string) {
			defer wg.Done()

			checkAvailabilityForTld(tldChan, availableChan)
		}(tldChan)
	}

	wg.Wait()
	close(availableChan)
}

func checkAvailabilityForTld(tldChan chan string, availableChan chan string) {
	for domain := range tldChan {
		log.Printf("processing domain: %s\n", domain)
		time.Sleep(1 * time.Second)
	}
}

func partitionBy(itemsChan chan string, partChanChan chan chan string, partitionKey func(string) string) {
	partChanMap := make(map[string]chan string)

	for item := range itemsChan {
		partKey := partitionKey(item)

		if partChan, ok := partChanMap[partKey]; ok {
			partChan <- item
		} else {
			partChan = make(chan string, 100)
			partChanMap[partKey] = partChan
			partChanChan <- partChan
			partChan <- item
		}
	}

	for _, v := range partChanMap {
		close(v)
	}
	close(partChanChan)
}
