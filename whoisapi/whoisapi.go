package whoisapi

import (
	"encoding/json"
	"fmt"
	"github.com/ali-l/domain_name_checker/config"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const layoutISO = "2006-01-02"

var baseURL = "https://api.jsonwhois.io/whois/domain?key=" + config.New().WhoisAPIKey + "&domain="

type whoisResponse struct {
	Result whoisResult
}

type whoisResult struct {
	Expires string
}

func (r whoisResponse) expiryDate() (time.Time, error) {
	t, err := time.Parse(layoutISO, strings.Split(r.Result.Expires, " ")[0])
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing time: %w", err)
	}

	return t, nil
}

func GetExpiry(domain string) (time.Time, error) {
	res, err := http.Get(baseURL + domain)
	if err != nil {
		return time.Time{}, fmt.Errorf("error getting response: %w", err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return time.Time{}, fmt.Errorf("error reading response body: %w", err)
	}

	response := whoisResponse{}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing response %w", err)
	}

	return response.expiryDate()
}