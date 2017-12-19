const https = require('https');
const AWS = require('aws-sdk');
const ses = new AWS.SES();

const BASE_URL = 'https://www.cineplex.com/Movie/';
const EMAIL_ADDRESS = '';
const MOVIES = [];

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

function checkSaleStatus(movie) {
  return new Promise((resolve, reject) => {
    console.log('Checking ' + movie);

    fetch(BASE_URL + movie)
      .then((html) => {
        let onSale = html.includes('quick-tickets');
        console.log(movie + ' ' + onSale);

        resolve(onSale)
      })
      .catch(err => {
        console.log(err);
        reject(err)
      })
  })
}

function emailIfOnSale(movies) {
  return new Promise((_, reject) => {
    movies.forEach((movie) => {
      checkSaleStatus(movie)
        .then(onSale => {
          if (onSale) {
            sendEmail({
              to: EMAIL_ADDRESS,
              subject: `Cineplex tickets on sale for ${movie}`,
              body: `${BASE_URL}${movie}`
            })
          }
        })
        .catch((err) => {
          sendEmail({
            to: EMAIL_ADDRESS,
            subject: `Cineplex checker failed for ${movie}`,
            body: `${err} ${BASE_URL}${movie}`
          });

          reject(err)
        })
    })
  })
}

exports.handler = (event, context, callback) => {
  emailIfOnSale(MOVIES)
    .catch(callback)
};