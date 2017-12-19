const https = require('https');
const AWS = require('aws-sdk');
const ses = new AWS.SES();

const API_KEY = '';
const BASE_URL = `https://api.jsonwhois.io/whois/domain?key=${API_KEY}&domain=`;
const EMAIL_ADDRESS = '';
const NOTIFY_THRESHOLD = 7776000000;
const DOMAINS = [];

function fetch(url) {
  return new Promise((resolve, reject) => {
    https
      .get(url, resp => {
        const { statusCode } = resp;
        if (statusCode !== 200) reject(`Status code: ${statusCode}`);

        let data = '';
        resp.on('data', chunk => data += chunk);
        resp.on('end', () => resolve(data));
      })
      .on('error', reject)
  })
}

function sendEmail({ to, subject, body }) {
  ses.sendEmail({
      Source: to,
      Destination: { ToAddresses: [to] },
      Message: {
        Subject: {
          Data: subject
        },
        Body: {
          Text: {
            Data: body
          }
        }
      }
    }, (err, _data) => {
      if (err) throw err;
    });
}

function getExpiryDate(domain) {
  return new Promise((resolve, reject) => {
    console.log('Checking ' + domain);

    fetch(BASE_URL + domain)
      .then((responseString) => {
        let {result: {expires}} = JSON.parse(responseString);

        console.log(domain + ' ' + expires);

        resolve(new Date(expires))
      })
      .catch(err => {
        console.log(err);
        reject(err)
      })
  })
}

function shouldNotify(date) {
  return date - new Date() < NOTIFY_THRESHOLD
}

function notifyIfNeeded(domains) {
  return new Promise((_, reject) => {
    domains.forEach((domain) => {
      getExpiryDate(domain)
        .then(date => {
          if (shouldNotify(date)) {
            sendEmail({
              to: EMAIL_ADDRESS,
              subject: `Domain name ${domain} expiring on ${date}`,
              body: `${domain}`
            })
          }
        })
        .catch((err) => {
          sendEmail({
            to: EMAIL_ADDRESS,
            subject: `Domain name check failed for ${domain}`,
            body: `${err} ${domain}`
          });

          reject(err)
        })
    })
  })
}

exports.handler = (event, context, callback) => {
  notifyIfNeeded(DOMAINS)
    .catch(callback)
};