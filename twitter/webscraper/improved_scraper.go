package webscraper

import (
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
)

// ImprovedSearcher uses Colly for better web scraping
type ImprovedSearcher struct {
	collector *colly.Collector
}

// NewImprovedSearcher creates a new improved searcher using Colly
func NewImprovedSearcher() *ImprovedSearcher {
	c := colly.NewCollector(
		colly.Debugger(&debug.LogDebugger{}),
	)

	// Set up realistic browser headers
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"

	// Set up rate limiting (be respectful)
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*duckduckgo*",
		Parallelism: 1,
		Delay:       2 * time.Second,
	})

	// Handle errors
	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Error: %v", err)
	})

	return &ImprovedSearcher{
		collector: c,
	}
}

// SearchWithColly performs a search using Colly
func (s *ImprovedSearcher) SearchWithColly(searchTerm string) ([]Article, error) {
	var articles []Article
	var scrapeError error

	// Clone collector for this search
	c := s.collector.Clone()

	// Set up the scraping logic
	c.OnHTML(".result", func(e *colly.HTMLElement) {
		title := strings.TrimSpace(e.ChildText(".result__title a"))
		if title == "" {
			title = strings.TrimSpace(e.ChildText("a[data-testid='result-title-a']"))
		}

		link := e.ChildAttr(".result__title a", "href")
		if link == "" {
			link = e.ChildAttr("a[data-testid='result-title-a']", "href")
		}

		description := strings.TrimSpace(e.ChildText(".result__snippet"))
		if description == "" {
			description = strings.TrimSpace(e.ChildText("[data-testid='result-snippet']"))
		}

		if title != "" && link != "" {
			// Parse redirect URL if needed
			actualURL := extractActualURL(link)
			if actualURL == "" {
				actualURL = link
			}

			articles = append(articles, Article{
				Title:       cleanText(title),
				URL:         actualURL,
				Description: cleanText(description),
			})
		}
	})

	c.OnResponse(func(r *colly.Response) {
		log.Printf("Response status: %d for URL: %s", r.StatusCode, r.Request.URL)
	})

	// Build search URL
	searchURL := buildDuckDuckGoURL(searchTerm)
	log.Printf("Searching URL: %s", searchURL)

	// Visit the search URL
	err := c.Visit(searchURL)
	if err != nil {
		scrapeError = fmt.Errorf("failed to visit search URL: %w", err)
	}

	// Wait for scraping to complete
	c.Wait()

	if scrapeError != nil {
		return nil, scrapeError
	}

	// Limit to top 10
	if len(articles) > 10 {
		articles = articles[:10]
	}

	return articles, nil
}

// SearchAsync performs asynchronous search with Colly
func (s *ImprovedSearcher) SearchAsync(searchTerm string, resultChan chan<- SearchResult, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("Starting Colly search for: %s", searchTerm)

	articles, err := s.SearchWithColly(searchTerm)

	result := SearchResult{
		SearchTerm: searchTerm,
		Articles:   articles,
		Error:      err,
	}

	resultChan <- result
	log.Printf("Completed Colly search for: %s, found %d articles", searchTerm, len(articles))
}

// Helper functions
func buildDuckDuckGoURL(searchTerm string) string {
	baseURL := "https://duckduckgo.com/html/"
	params := url.Values{}
	params.Add("q", searchTerm)
	params.Add("t", "h_")
	params.Add("ia", "web")

	return baseURL + "?" + params.Encode()
}

func extractActualURL(redirectURL string) string {
	if strings.HasPrefix(redirectURL, "/l/?uddg=") {
		parsedURL, err := url.Parse("https://duckduckgo.com" + redirectURL)
		if err != nil {
			return ""
		}

		uddg := parsedURL.Query().Get("uddg")
		if uddg != "" {
			decodedURL, err := url.QueryUnescape(uddg)
			if err != nil {
				return uddg
			}
			return decodedURL
		}
	}

	if strings.HasPrefix(redirectURL, "http") {
		return redirectURL
	}

	return ""
}

func cleanText(text string) string {
	re := regexp.MustCompile(`\s+`)
	cleaned := re.ReplaceAllString(text, " ")
	return strings.TrimSpace(cleaned)
}
