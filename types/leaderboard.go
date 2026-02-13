package types

import "github.com/ethereum/go-ethereum/common"

// Leaderboard represents leaderboard statistics for a user
type Leaderboard struct {
	AllocationRoundPoints float64 `json:"allocationRoundPoints"`
	TotalPoints           float64 `json:"totalPoints"`
	Rank                  int     `json:"rank"`
}

// LeaderboardAccount represents the account information returned by the leaderboard GraphQL query
type LeaderboardAccount struct {
	Name        string         `json:"name"`
	Address     common.Address `json:"-"`
	RawAddress  string         `json:"address"`
	ImageURL    *string        `json:"imageUrl"`
	ImageStatus string         `json:"imageStatus"`
	Leaderboard *Leaderboard   `json:"leaderboard"`
}

// GraphQLRequest represents a GraphQL request payload
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]any `json:"variables"`
	OperationName string                `json:"operationName"`
}

// GraphQLResponse represents a generic GraphQL response
type GraphQLResponse[T any] struct {
	Data   T               `json:"data"`
	Errors []GraphQLError  `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message string `json:"message"`
}

// GetLeaderboardUserStatsData represents the data field of GetLeaderboardUserStats response
type GetLeaderboardUserStatsData struct {
	Account *LeaderboardAccount `json:"account"`
}
