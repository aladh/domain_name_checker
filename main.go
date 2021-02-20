package main

import (
	"fmt"
	"github.com/ali-l/domain_name_checker/whoisapi"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

const notifyThreshold = 720 // 30 days

var domains []string

func init() {
	domains = strings.Split(os.Args[1], ",")
}

func main() {
	availableChan := make(chan string, len(domains))
	wg := sync.WaitGroup{}

	for _, domain := range domains {
		wg.Add(1)

		go func(domain string, availableChan chan<- string) {
			defer wg.Done()

			available, err := isAvailable(domain)
			if err != nil {
				log.Printf("error checking availability: %s\n", err)
				return
			}

			if available {
				log.Printf("Domain name %s is available\n", domain)
				availableChan <- domain
			}
		}(domain, availableChan)
	}

	wg.Wait()
	close(availableChan)

	if len(availableChan) > 0 {
		log.Fatalln("One or more domains is available!")
	}
}

func isAvailable(domain string) (bool, error) {
	log.Printf("Checking %s\n", domain)

	t, err := whoisapi.GetExpiry(domain)
	if err != nil {
		return false, fmt.Errorf("error getting expiry date for domain %s: %w", domain, err)
	}

	return time.Until(t) <= notifyThreshold*time.Hour, nil
}