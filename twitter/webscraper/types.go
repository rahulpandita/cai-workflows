package webscraper

// Article represents a search result article
type Article struct {
	Title       string
	URL         string
	Description string
}

// SearchResult represents the result of a search operation
type SearchResult struct {
	SearchTerm string
	Articles   []Article
	Error      error
}
