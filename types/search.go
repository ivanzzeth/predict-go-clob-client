package types

// ForwardPaginationInput represents GraphQL forward pagination input
type ForwardPaginationInput struct {
	First *int    `json:"first,omitempty"`
	After *string `json:"after,omitempty"`
}

// SearchFilterInput represents the filter input for search query
// status: CategoryStatus enum (OPEN, RESOLVED)
// tags: list of tag IDs
type SearchFilterInput struct {
	Status *CategoryStatus `json:"status,omitempty"`
	Tags   []string        `json:"tags,omitempty"` // tag IDs (GraphQL ID type)
}

// SearchOptions represents options for the Search query
type SearchOptions struct {
	Query      string                  // Required: search keyword
	Filter     *SearchFilterInput      // Optional: filter by status/tags
	Pagination *ForwardPaginationInput // Optional: pagination
}

// --- GraphQL Connection response types (Relay-style) ---

// SearchCategoryStatistics represents category statistics in search results
type SearchCategoryStatistics struct {
	VolumeTotalUsd float64 `json:"volumeTotalUsd"`
}

// SearchCategoryTag represents a tag in search results
type SearchCategoryTag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// SearchCategoryTagEdge represents an edge in the category tag connection
type SearchCategoryTagEdge struct {
	Node SearchCategoryTag `json:"node"`
}

// SearchCategoryTagConnection represents a connection of category tags
type SearchCategoryTagConnection struct {
	Edges []SearchCategoryTagEdge `json:"edges"`
}

// SearchCategory represents a category in search results
type SearchCategory struct {
	ID         string                      `json:"id"` // slug-style string ID
	Title      string                      `json:"title"`
	ImageURL   string                      `json:"imageUrl"`
	EndsAt     *string                     `json:"endsAt,omitempty"` // nullable datetime string
	Statistics SearchCategoryStatistics    `json:"statistics"`
	Tags       SearchCategoryTagConnection `json:"tags"`
}

// SearchCategoryEdge represents an edge in the category connection
type SearchCategoryEdge struct {
	Node SearchCategory `json:"node"`
}

// SearchCategoryConnection represents a connection of categories
type SearchCategoryConnection struct {
	Edges []SearchCategoryEdge `json:"edges"`
}

// SearchMarketStatistics represents market statistics in search results
type SearchMarketStatistics struct {
	VolumeTotalUsd float64 `json:"volumeTotalUsd"`
}

// SearchOutcome represents an outcome in search results
type SearchOutcome struct {
	ID    string `json:"id"`
	Index int    `json:"index"`
	Name  string `json:"name"`
}

// SearchOutcomeEdge represents an edge in the outcome connection
type SearchOutcomeEdge struct {
	Node SearchOutcome `json:"node"`
}

// SearchOutcomeConnection represents a connection of outcomes
type SearchOutcomeConnection struct {
	Edges []SearchOutcomeEdge `json:"edges"`
}

// SearchMarketCategory represents the category info embedded in a market search result
type SearchMarketCategory struct {
	ID   string                      `json:"id"`
	Tags SearchCategoryTagConnection `json:"tags"`
}

// SearchMarket represents a market in search results
type SearchMarket struct {
	ID               string                  `json:"id"`
	ImageURL         string                  `json:"imageUrl"`
	Question         string                  `json:"question"`
	DecimalPrecision int                     `json:"decimalPrecision"`
	Statistics       SearchMarketStatistics  `json:"statistics"`
	Outcomes         SearchOutcomeConnection `json:"outcomes"`
	Category         SearchMarketCategory    `json:"category"`
}

// SearchMarketEdge represents an edge in the market connection
type SearchMarketEdge struct {
	Cursor string       `json:"cursor"`
	Node   SearchMarket `json:"node"`
}

// SearchMarketConnection represents a connection of markets
type SearchMarketConnection struct {
	Edges []SearchMarketEdge `json:"edges"`
}

// SearchResult represents the search result containing categories and markets
type SearchResult struct {
	Categories SearchCategoryConnection `json:"categories"`
	Markets    SearchMarketConnection   `json:"markets"`
}

// SearchData represents the data field of the Search GraphQL response
type SearchData struct {
	Search SearchResult `json:"search"`
}
