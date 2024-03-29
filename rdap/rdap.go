package rdap

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/aladh/domain_name_checker/domain"
)

type rdapResponse struct {
	Events []rdapEvent `json:"events"`
}

type rdapEvent struct {
	Action string `json:"eventAction"`
	Date   string `json:"eventDate"`
}

const expirationAction = "expiration"

func Initialize() error {
	err := bootstrapServiceRegistry()
	if err != nil {
		return fmt.Errorf("error bootstrapping service registry: %w", err)
	}

	return nil
}

func ExpiryDate(domain *domain.Domain) (string, error) {
	serviceURL, err := serviceForTld(domain.Tld)
	if err != nil {
		return "", fmt.Errorf("error getting service for domain %s", domain.Name)
	}

	resp, err := http.Get(fmt.Sprintf("%sdomain/%s", serviceURL, domain.Name))
	if err != nil {
		return "", fmt.Errorf("error making request for domain %s", domain.Name)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("received bad response code: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return "", fmt.Errorf("error reading response body %w", err)
	}

	var rdapRes rdapResponse
	err = json.Unmarshal(respBody, &rdapRes)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling JSON response: %w", err)
	}

	for _, event := range rdapRes.Events {
		if event.Action == expirationAction {
			return event.Date, nil
		}
	}

	return "", fmt.Errorf("unable to find expiration action for domain %s", domain)
}
