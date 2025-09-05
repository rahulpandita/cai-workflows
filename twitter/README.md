# Twitter Following Tweets Fetcher

A Go application that fetches the latest 100 tweets from accounts you follow on Twitter.

## Prerequisites

1. **Twitter Developer Account**: You need to have a Twitter Developer account and create an app to get API credentials.
   - Visit https://developer.twitter.com/
   - Create a new app
   - Generate your API keys and tokens

2. **Go 1.21+**: Make sure you have Go installed on your system.

## Setup

1. **Install dependencies**:
   ```bash
   cd /workspaces/cai-workflows/twitter
   go mod tidy
   ```

2. **Set up environment variables**:
   ```bash
   cp .env.example .env
   ```
   
   Edit the `.env` file and add your Twitter API credentials:
   ```bash
   export TWITTER_CONSUMER_KEY="your_consumer_key_here"
   export TWITTER_CONSUMER_SECRET="your_consumer_secret_here"
   export TWITTER_ACCESS_TOKEN="your_access_token_here"
   export TWITTER_ACCESS_SECRET="your_access_secret_here"
   ```

3. **Load environment variables**:
   ```bash
   source .env
   ```

## Usage

Run the application directly:
```bash
go run main.go
```

Or build and run:
```bash
go build -o twitter-fetcher
./twitter-fetcher
```

Or use the Makefile:
```bash
make run-with-env
```

The application will:
1. Prompt you for your Twitter username and password
2. Fetch your following list
3. Retrieve the latest tweets from each account you follow
4. Display up to 100 tweets on the console

## Features

- Secure password input (hidden typing)
- Fetches tweets from your following list
- Displays tweet content, creation time, likes, and retweets
- Limits output to 100 tweets total
- Excludes replies and retweets for cleaner output

## Important Notes

- **API Limits**: Twitter API has rate limits. The app respects these limits.
- **Authentication**: Uses OAuth 1.0a for Twitter API authentication.
- **Privacy**: Your username and password are only used locally and not stored.

## Troubleshooting

1. **API Rate Limits**: If you hit rate limits, wait 15 minutes before running again.
2. **Invalid Credentials**: Double-check your API keys and tokens.
3. **Private Accounts**: The app can only fetch tweets from public accounts.

## Dependencies

- `github.com/dghubble/go-twitter`: Twitter API client
- `github.com/dghubble/oauth1`: OAuth 1.0a implementation
- `golang.org/x/term`: For secure password input
