# Domain Name Checker
Sends you an email when a domain name is going to expire within the next month

### Configuration
Deployed as a Cron triggered AWS Lambda

- Set environment variable API_KEY with key for [JsonWhois](https://jsonwhois.io/) API
- Set environment variable EMAIL_ADDRESS with email to notify
- Fill in DOMAINS constant in index.js with a list of domain names (ex. abc.com) to check