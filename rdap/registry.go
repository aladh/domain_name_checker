package rdap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const bootstrapUrl = "https://data.iana.org/rdap/dns.json"

var reg registry

type registry struct {
	Services [][][]string `json:"services"`
}

func serviceForTld(lookupTld string) (string, error) {
	for _, tldServicePair := range reg.Services {
		for _, tld := range tldServicePair[0] {
			if tld == lookupTld {
				return tldServicePair[1][0], nil
			}
		}
	}

	return "", fmt.Errorf("couldn't find service for tld %s", lookupTld)
}

func bootstrapServiceRegistry() error {
	resp, err := http.Get(bootstrapUrl)
	if err != nil {
		return fmt.Errorf("error making request for RDAP bootstrap file: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("received bad response code: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error reading response body %w", err)
	}

	err = json.Unmarshal(respBody, &reg)
	if err != nil {
		return fmt.Errorf("error unmarshalling JSON response: %w", err)
	}

	return nil
}
