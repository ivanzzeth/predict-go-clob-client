package predictclob

import (
	"encoding/json"
	"fmt"

	"github.com/ivanzzeth/predict-go-clob-client/types"
)

const searchQuery = `query SearchQuery($input: String!, $filter: SearchFilterInput, $pagination: ForwardPaginationInput) {
  search(query: $input, filter: $filter, pagination: $pagination) {
    categories {
      edges {
        node {
          ...CategorySearchResult
        }
      }
    }
    markets {
      edges {
        cursor
        node {
          ...MarketSearchResult
        }
      }
    }
  }
}

fragment CategorySearchResult on Category {
  id
  title
  imageUrl
  endsAt
  statistics {
    volumeTotalUsd
  }
  tags {
    edges {
      node {
        id
        name
      }
    }
  }
}

fragment MarketSearchResult on Market {
  id
  imageUrl
  question
  decimalPrecision
  statistics {
    volumeTotalUsd
  }
  outcomes {
    edges {
      node {
        id
        index
        name
      }
    }
  }
  category {
    id
    tags {
      edges {
        node {
          id
          name
        }
      }
    }
  }
}`

// Search queries the GraphQL API to search for categories and markets.
// This uses the GraphQL API endpoint (graphql.predict.fun).
// No authentication required.
func (c *Client) Search(opts *types.SearchOptions) (*types.SearchResult, error) {
	if opts == nil || opts.Query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	variables := map[string]any{
		"input": opts.Query,
	}

	if opts.Filter != nil {
		variables["filter"] = opts.Filter
	}

	if opts.Pagination != nil {
		variables["pagination"] = opts.Pagination
	}

	request := &types.GraphQLRequest{
		Query:         searchQuery,
		Variables:     variables,
		OperationName: "SearchQuery",
	}

	respBody, err := c.doGraphQLRequest(request)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}

	var response types.GraphQLResponse[types.SearchData]
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	if len(response.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL errors: %s", response.Errors[0].Message)
	}

	return &response.Data.Search, nil
}
