package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/dghubble/oauth1"
)

type TwitterConfig struct {
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string
}

type TwitterUser struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Username      string `json:"username"`
	Description   string `json:"description"`
	Location      string `json:"location"`
	CreatedAt     string `json:"created_at"`
	PublicMetrics struct {
		FollowersCount int `json:"followers_count"`
		FollowingCount int `json:"following_count"`
		TweetCount     int `json:"tweet_count"`
	} `json:"public_metrics"`
}

type Tweet struct {
	ID            string `json:"id"`
	Text          string `json:"text"`
	CreatedAt     string `json:"created_at"`
	AuthorID      string `json:"author_id"`
	PublicMetrics struct {
		RetweetCount int `json:"retweet_count"`
		LikeCount    int `json:"like_count"`
		ReplyCount   int `json:"reply_count"`
		QuoteCount   int `json:"quote_count"`
	} `json:"public_metrics"`
}

type TwitterAPIResponse struct {
	Data     []Tweet     `json:"data"`
	Meta     interface{} `json:"meta"`
	Includes *struct {
		Users []TwitterUser `json:"users"`
	} `json:"includes,omitempty"`
}

type UserAPIResponse struct {
	Data TwitterUser `json:"data"`
}

type TwitterClient struct {
	httpClient *http.Client
	baseURL    string
}

func main() {
	fmt.Println("Twitter Following Tweets Fetcher")
	fmt.Println("=================================")

	config := getTwitterAPICredentials()
	twitterClient := initializeTwitterClient(config)

	fmt.Println("\nGetting your account information...")
	user, err := twitterClient.getAuthenticatedUser()
	if err != nil {
		log.Fatalf("Error getting your account info: %v", err)
	}

	fmt.Printf("Logged in as: @%s (%s)\n", user.Username, user.Name)
	fmt.Printf("Following: %d accounts\n", user.PublicMetrics.FollowingCount)

	fmt.Println("\nSearching for tweets from popular tech accounts...")
	err = twitterClient.searchRecentTweets()
	if err != nil {
		fmt.Println("\nFalling back to your recent tweets as example:")
		twitterClient.fetchUserTweets(user.ID)
	}
}

func getTwitterAPICredentials() TwitterConfig {
	config := TwitterConfig{
		ConsumerKey:    os.Getenv("TWITTER_CONSUMER_KEY"),
		ConsumerSecret: os.Getenv("TWITTER_CONSUMER_SECRET"),
		AccessToken:    os.Getenv("TWITTER_ACCESS_TOKEN"),
		AccessSecret:   os.Getenv("TWITTER_ACCESS_SECRET"),
	}

	if config.ConsumerKey == "" || config.ConsumerSecret == "" ||
		config.AccessToken == "" || config.AccessSecret == "" {
		fmt.Println("Please set environment variables for Twitter API credentials")
		os.Exit(1)
	}
	return config
}

func initializeTwitterClient(config TwitterConfig) *TwitterClient {
	oauthConfig := oauth1.NewConfig(config.ConsumerKey, config.ConsumerSecret)
	token := oauth1.NewToken(config.AccessToken, config.AccessSecret)
	httpClient := oauthConfig.Client(oauth1.NoContext, token)

	return &TwitterClient{
		httpClient: httpClient,
		baseURL:    "https://api.twitter.com/2",
	}
}

func (tc *TwitterClient) getAuthenticatedUser() (*TwitterUser, error) {
	url := fmt.Sprintf("%s/users/me?user.fields=public_metrics", tc.baseURL)

	resp, err := tc.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var response UserAPIResponse
	json.Unmarshal(body, &response)
	return &response.Data, nil
}

func (tc *TwitterClient) fetchUserTweets(userID string) error {
	url := fmt.Sprintf("%s/users/%s/tweets?max_results=5&tweet.fields=created_at,public_metrics", tc.baseURL, userID)

	resp, err := tc.httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var response TwitterAPIResponse
	json.Unmarshal(body, &response)

	fmt.Printf("\n--- Your Recent Tweets ---\n")
	for i, tweet := range response.Data {
		fmt.Printf("\n%d. %s\n", i+1, tweet.Text)
		fmt.Printf("   Likes: %d | Retweets: %d\n",
			tweet.PublicMetrics.LikeCount,
			tweet.PublicMetrics.RetweetCount)
	}
	return nil
}

func (tc *TwitterClient) searchRecentTweets() error {
	searchTerms := []string{
		"from:github -is:retweet",
		"golang programming -is:retweet",
	}

	totalTweets := 0

	for _, searchTerm := range searchTerms {
		if totalTweets >= 10 {
			break
		}

		encodedQuery := url.QueryEscape(searchTerm)
		apiURL := fmt.Sprintf("%s/tweets/search/recent?query=%s&max_results=5&tweet.fields=created_at,public_metrics,author_id&expansions=author_id&user.fields=username,name",
			tc.baseURL, encodedQuery)

		fmt.Printf("\nSearching for: %s\n", searchTerm)

		resp, err := tc.httpClient.Get(apiURL)
		if err != nil {
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == 429 {
			fmt.Println("⚠️  Rate limit reached (Free tier: 1 search per 15 minutes)")
			return fmt.Errorf("rate limit")
		}

		if resp.StatusCode != 200 {
			fmt.Printf("API error (status: %d)\n", resp.StatusCode)
			continue
		}

		var response TwitterAPIResponse
		if err := json.Unmarshal(body, &response); err != nil {
			continue
		}

		if len(response.Data) == 0 {
			fmt.Printf("No tweets found for: %s\n", searchTerm)
			continue
		}

		// Create user map
		userMap := make(map[string]TwitterUser)
		if response.Includes != nil && response.Includes.Users != nil {
			for _, user := range response.Includes.Users {
				userMap[user.ID] = user
			}
		}

		fmt.Printf("--- Found %d tweets ---\n", len(response.Data))

		for _, tweet := range response.Data {
			if totalTweets >= 10 {
				break
			}

			authorInfo := "Unknown"
			if author, exists := userMap[tweet.AuthorID]; exists {
				authorInfo = fmt.Sprintf("@%s", author.Username)
			}

			fmt.Printf("\n%d. %s\n", totalTweets+1, authorInfo)
			fmt.Printf("   %s\n", tweet.Text)
			fmt.Printf("   Likes: %d | Retweets: %d\n",
				tweet.PublicMetrics.LikeCount,
				tweet.PublicMetrics.RetweetCount)

			totalTweets++
		}

		time.Sleep(2 * time.Second) // Rate limiting
	}

	if totalTweets == 0 {
		return fmt.Errorf("no tweets found")
	}

	fmt.Printf("\nTotal tweets found: %d\n", totalTweets)
	fmt.Println("\n💡 Note: To get tweets from YOUR actual following list, upgrade to Basic/Pro tier")
	return nil
}
