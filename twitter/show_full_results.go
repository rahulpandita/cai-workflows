package main

import (
	"fmt"
	"strings"
	"sync"

	"twitter-follower-tweets/webscraper"
)

func main() {
	fmt.Println("=== Full Results from Colly Scraper ===")

	searcher := webscraper.NewSearcher()

	var wg sync.WaitGroup
	resultChan := make(chan webscraper.SearchResult, 1)

	searchTerm := "golang programming"

	wg.Add(1)
	go searcher.SearchAsync(searchTerm, resultChan, &wg)

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		if result.Error != nil {
			fmt.Printf("❌ Error: %v\n", result.Error)
			return
		}

		fmt.Printf("✅ Found %d articles for '%s':\n\n", len(result.Articles), result.SearchTerm)

		for i, article := range result.Articles {
			fmt.Printf("📄 %d. %s\n", i+1, article.Title)
			fmt.Printf("🔗 URL: %s\n", article.URL)
			if article.Description != "" {
				fmt.Printf("📝 Description: %s\n", article.Description)
			}
			fmt.Println(strings.Repeat("-", 80))
		}
	}

	fmt.Println("\n🎉 Success! The Colly-based scraper works much better!")
	fmt.Println("\n💡 Key improvements:")
	fmt.Println("- Better bot detection avoidance")
	fmt.Println("- Built-in rate limiting")
	fmt.Println("- More robust error handling")
	fmt.Println("- Better CSS selector handling")
}
