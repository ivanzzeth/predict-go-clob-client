package types

import (
	"encoding/json"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// CategorySort represents the sort order for categories
type CategorySort string

const (
	CategorySortVolume24HDesc  CategorySort = "VOLUME_24H_DESC"
	CategorySortVolumeAllDesc   CategorySort = "VOLUME_ALL_DESC"
	CategorySortPublishedAtAsc  CategorySort = "PUBLISHED_AT_ASC"
	CategorySortPublishedAtDesc CategorySort = "PUBLISHED_AT_DESC"
)

// String returns the string representation
func (s CategorySort) String() string {
	return string(s)
}

// IsValid checks if the sort order is valid
func (s CategorySort) IsValid() bool {
	return s == CategorySortVolume24HDesc || s == CategorySortVolumeAllDesc ||
		s == CategorySortPublishedAtAsc || s == CategorySortPublishedAtDesc
}

// MarketVariant represents the variant of a market
type MarketVariant string

const (
	MarketVariantDefault      MarketVariant = "DEFAULT"
	MarketVariantSportsMatch  MarketVariant = "SPORTS_MATCH"
	MarketVariantCryptoUpDown MarketVariant = "CRYPTO_UP_DOWN"
	MarketVariantTweetCount   MarketVariant = "TWEET_COUNT"
)

// String returns the string representation
func (v MarketVariant) String() string {
	return string(v)
}

// IsValid checks if the variant is valid
func (v MarketVariant) IsValid() bool {
	return v == MarketVariantDefault || v == MarketVariantSportsMatch ||
		v == MarketVariantCryptoUpDown || v == MarketVariantTweetCount
}

// CategoryTag represents a tag associated with a category
type CategoryTag struct {
	ID   IntegerOrString `json:"id"`   // bigint string
	Name string          `json:"name"`
}

// CategoryMarket represents a market within a category
// This is a more complete representation than the base Market type
type CategoryMarket struct {
	ID                     MarketID              `json:"id"`
	ImageURL               string                `json:"imageUrl"`
	Title                  string                `json:"title"`
	Question               string                `json:"question"`
	Description            string                `json:"description"`
	Status                 MarketStatus          `json:"status"`
	IsNegRisk              bool                  `json:"isNegRisk"`
	IsYieldBearing         bool                  `json:"isYieldBearing"`
	FeeRateBps             FeeRateBps            `json:"feeRateBps"`
	Resolution             *CategoryMarketResolution `json:"resolution,omitempty"` // nullable
	OracleQuestionID       string                `json:"oracleQuestionId"`
	ConditionID            string                `json:"conditionId"`
	ResolverAddress        common.Address        `json:"resolverAddress"`
	Outcomes               []CategoryMarketOutcome `json:"outcomes"`
	QuestionIndex          *int                  `json:"questionIndex,omitempty"` // nullable
	SpreadThreshold        float64               `json:"spreadThreshold"`
	ShareThreshold         float64               `json:"shareThreshold"`
	PolymarketConditionIDs []string              `json:"polymarketConditionIds"`
	KalshiMarketTicker     *string               `json:"kalshiMarketTicker,omitempty"` // nullable
	CategorySlug           string                `json:"categorySlug"`
	CreatedAt              time.Time            `json:"createdAt"`
	DecimalPrecision       int                   `json:"decimalPrecision"`
}

// CategoryMarketResolution represents the resolution of a market in a category
type CategoryMarketResolution struct {
	Name      string       `json:"name"`
	IndexSet  int          `json:"indexSet"`
	OnChainID TokenID      `json:"onChainId"`
	Status    *OutcomeStatus `json:"status,omitempty"` // nullable
}

// CategoryMarketOutcome represents an outcome in a category market
type CategoryMarketOutcome struct {
	Name      string       `json:"name"`
	IndexSet  int          `json:"indexSet"`
	OnChainID TokenID      `json:"onChainId"`
	Status    *OutcomeStatus `json:"status,omitempty"` // nullable
}

// OutcomeStatus is defined in match.go

// Category represents a market category with all fields from the API
type Category struct {
	ID             IntegerOrString  `json:"id"`             // number, converted via IntegerOrString
	Slug           string          `json:"slug"`
	Title          string          `json:"title"`
	Description    string          `json:"description"`
	ImageURL       string          `json:"imageUrl"`
	IsNegRisk      bool            `json:"isNegRisk"`
	IsYieldBearing bool            `json:"isYieldBearing"`
	MarketVariant  MarketVariant    `json:"marketVariant"`
	CreatedAt      time.Time       `json:"createdAt"`
	PublishedAt    time.Time       `json:"publishedAt"`
	Markets        []CategoryMarket `json:"markets"`
	StartsAt       time.Time        `json:"startsAt"`
	EndsAt          *time.Time      `json:"endsAt,omitempty"` // nullable
	Status          CategoryStatus  `json:"status"`
	Tags            []CategoryTag   `json:"tags"`
}

// UnmarshalJSON implements custom unmarshaling for Category
func (c *Category) UnmarshalJSON(data []byte) error {
	type Alias Category
	aux := &struct {
		ID interface{} `json:"id"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Convert ID using IntegerOrString
	var ios IntegerOrString
	idBytes, err := json.Marshal(aux.ID)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(idBytes, &ios); err != nil {
		return err
	}
	c.ID = ios

	return nil
}

// GetCategoriesOptions represents options for getting categories
type GetCategoriesOptions struct {
	First  *int          `json:"first,omitempty"`  // string to be decoded into a number
	After  *string       `json:"after,omitempty"`  // pagination cursor
	Status CategoryStatus `json:"status,omitempty"` // OPEN or RESOLVED
	Sort   CategorySort  `json:"sort,omitempty"`   // sort order
}

// GetCategoriesResponse represents the response from GetCategories
type GetCategoriesResponse struct {
	Success bool       `json:"success"`
	Cursor  *string    `json:"cursor,omitempty"` // nullable
	Data    []Category `json:"data"`
}
