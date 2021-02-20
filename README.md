# Domain Name Checker

Notifies you when a domain is available or expiring soon.

### Configuration

- Set environment variable WHOIS_API_KEY with key for [JsonWhois](https://jsonwhois.io/) API
- Deployed as a scheduled job on GitLab CI.

#### Build

```shell
go build
```

#### Run

```shell
WHOIS_API_KEY=KEY ./domain_name_checker example.com 
```
