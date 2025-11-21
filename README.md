# DevZen Shownote Generator

![DevZen themes](https://raw.githubusercontent.com/SBozhko/devzen-shownote-generator/master/dz_2.png)
![Card to start recording](https://raw.githubusercontent.com/SBozhko/devzen-shownote-generator/master/dz_1.png)
![A theme with links](https://raw.githubusercontent.com/SBozhko/devzen-shownote-generator/master/dz_3.png)

## Tech Stack

This service is written in **Go** (rewritten from Scala in 2025) using only the standard library.

## Building and Running Locally

```bash
# Build
go build -o devzen-shownote-generator ./cmd/server

# Run (requires environment variables - see below)
./devzen-shownote-generator

# Run tests
go test ./...
```

## Secrets for Heroku / Fly.io

On both platforms secrets are supplied via environment variables.

- https://fly.io/docs/flyctl/secrets/
- https://devcenter.heroku.com/articles/config-vars

### API keys and Tokens

- `TRELLO_APP_KEY`
- `TRELLO_READ_TOKEN`
- `TELEGRAM_BOT_TOKEN`

### Other IDs

- `TRELLO_DISCUSSED_LIST_ID`
- `TRELLO_TO_DISCUSS_LIST_ID`
- `TRELLO_IN_DISCUSSION_LIST_ID`
- `TRELLO_RECORDING_STARTED_CARD_ID`
- `TRELLO_BACKLOG_LIST_ID`

### Optional

- `PORT` (default: 9025)

## How to get Trello keys

#### TRELLO_APP_KEY 
Get Trello API Key from https://trello.com/app-key

#### TRELLO_READ_TOKEN
Get from: https://trello.com/1/authorize?expiration=never&name=DevZenShownotesGen&key=TRELLO_APP_KEY&scope=read&response_type=token

## How to register a webhook in Trello

```bash
curl -X POST -H "Content-Type: application/json" \
https://api.trello.com/1/tokens/{TRELLO_READ_TOKEN}/webhooks/ \
-d '{
  "key": "{TRELLO_APP_KEY}",
  "callbackURL": "https://{your_heroku_app_domain}/trellohook",
  "idModel":"{TRELLO_IN_DISCUSSION_LIST_ID}",
  "description": "DevZen_Webhook"
}'
```

Read more: https://developer.atlassian.com/cloud/trello/guides/rest-api/webhooks/

## How to get TELEGRAM_BOT_TOKEN

https://core.telegram.org/bots#creating-a-new-bot

## Deploying this app to Fly.io

To create an app for the first time if you don't have it follow [this guide](https://fly.io/docs/languages-and-frameworks/dockerfile/).

For subsequent deploys:

- Install [`flyctl`](https://fly.io/docs/getting-started/installing-flyctl/) - you'll need it.
- Login with `flyctl auth login`.
- Run `flyctl deploy`
