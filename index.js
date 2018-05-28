const { get } = require('https');
// noinspection NpmUsedModulesInstalled
const SES = require('aws-sdk/clients/ses');
// noinspection JSUnresolvedFunction
const ses = new SES();

const BASE_URL = `https://api.jsonwhois.io/whois/domain?key=${process.env.API_KEY}&domain=`;
const NOTIFY_THRESHOLD = 2592000000; // 30 days
const DOMAINS = [];

function fetch(url) {
  return new Promise((resolve, reject) => {
    get(url, (res) => {
      const { statusCode } = res;

      if (statusCode !== 200) {
        reject(`Bad status code: ${statusCode}`);
        res.destroy();
        return
      }

      let data = '';
      res.on('data', (chunk) => data += chunk);
      res.on('end', () => resolve(data))
    })
      .on('error', reject)
  })
}

function sendEmail({ to, subject, body }) {
  // noinspection JSUnresolvedFunction, JSUnusedLocalSymbols
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
    if (err) throw err
  })
}

async function expiryDate(domain) {
  const { result: { expires } } = JSON.parse(await fetch(BASE_URL + domain));
  return Date.parse(expires)
}

function expiringSoon(date) {
  return date - Date.now() < NOTIFY_THRESHOLD
}

exports.handler = () => {
  DOMAINS.forEach(async (domain) => {
    try {
      const date = await expiryDate(domain);

      expiringSoon(date) && sendEmail({
        to: process.env.EMAIL_ADDRESS,
        subject: `Domain name ${domain} expiring on ${new Date(date)}`,
        body: `${domain}`
      })
    } catch (e) {
      sendEmail({
        to: process.env.EMAIL_ADDRESS,
        subject: `Domain name check failed for ${domain}`,
        body: `${e} ${domain}`
      })
    }
  })
};