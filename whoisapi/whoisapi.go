package whoisapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/ali-l/domain_name_checker/config"
)

const layoutISO = "2006-01-02"

var baseURL = "https://api.jsonwhois.io/whois/domain?key=" + config.New().WhoisAPIKey + "&domain="

type whoisResponse struct {
	Result WhoisResult
}

type WhoisResult struct {
	Registered bool
	Expires    string
}

func (r WhoisResult) ExpirationDate() (time.Time, error) {
	t, err := time.Parse(layoutISO, strings.Split(r.Expires, " ")[0])
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing time: %w", err)
	}

	return t, nil
}

func GetWhoisInfo(domain string) (WhoisResult, error) {
	res, err := http.Get(baseURL + domain)
	if err != nil {
		return WhoisResult{}, fmt.Errorf("error getting response: %w", err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return WhoisResult{}, fmt.Errorf("error reading response body: %w", err)
	}

	response := whoisResponse{}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return WhoisResult{}, fmt.Errorf("error parsing response %w", err)
	}

	return response.Result, nil
}
