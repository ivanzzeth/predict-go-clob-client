package predictclob

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/ivanzzeth/predict-go-clob-client/types"
)

const getLeaderboardUserStatsQuery = `query GetLeaderboardUserStats($address: Address!) {
  account(address: $address) {
    name
    address
    imageUrl
    imageStatus
    leaderboard {
      allocationRoundPoints
      totalPoints
      rank
    }
  }
}`

// doGraphQLRequest sends a GraphQL request to the GraphQL endpoint.
// Unlike doRequest which targets the REST API, this targets the GraphQL API host.
func (c *Client) doGraphQLRequest(request *types.GraphQLRequest) ([]byte, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GraphQL request: %w", err)
	}

	url := c.graphqlHost + constants.EndpointGraphQL

	headers := map[string]string{
		"accept":       "application/graphql-response+json, application/json",
		"Content-Type": "application/json",
		"origin":       "https://predict.fun",
		"referer":      "https://predict.fun/",
	}

	resp, err := DoReqClientRequest(c.reqClient, http.MethodPost, url, body, headers, true)
	if err != nil {
		return nil, fmt.Errorf("GraphQL request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GraphQL API error (%d): %s", resp.StatusCode, resp.String())
	}

	return resp.Bytes(), nil
}

// GetLeaderboardUserStats queries leaderboard points and rank for a given address.
// This uses the GraphQL API endpoint (graphql.predict.fun).
// No authentication required.
func (c *Client) GetLeaderboardUserStats(address common.Address) (*types.LeaderboardAccount, error) {
	request := &types.GraphQLRequest{
		Query: getLeaderboardUserStatsQuery,
		Variables: map[string]any{
			"address": address.Hex(),
		},
		OperationName: "GetLeaderboardUserStats",
	}

	respBody, err := c.doGraphQLRequest(request)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaderboard user stats: %w", err)
	}

	var response types.GraphQLResponse[types.GetLeaderboardUserStatsData]
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse leaderboard response: %w", err)
	}

	if len(response.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL errors: %s", response.Errors[0].Message)
	}

	if response.Data.Account == nil {
		return nil, fmt.Errorf("account not found for address: %s", address.Hex())
	}

	// Convert RawAddress to common.Address
	account := response.Data.Account
	if account.RawAddress != "" {
		account.Address = common.HexToAddress(account.RawAddress)
	}

	return account, nil
}
