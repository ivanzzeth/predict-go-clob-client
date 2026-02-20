package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	predictclob "github.com/ivanzzeth/predict-go-clob-client"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/ivanzzeth/predict-go-clob-client/types"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file from project root directory
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get API key from environment (optional for this endpoint)
	apiKey := os.Getenv("PREDICT_API_KEY")

	// Create read-only client
	client := predictclob.NewReadOnlyClient(constants.DefaultAPIHost, apiKey)

	// SEARCH_QUERY: search keyword (required, default "btc")
	query := os.Getenv("SEARCH_QUERY")
	if query == "" {
		query = "btc"
	}

	// Build search options
	opts := &types.SearchOptions{
		Query: query,
	}

	// SEARCH_STATUS: filter by status (optional, e.g. "OPEN", "RESOLVED")
	if statusStr := os.Getenv("SEARCH_STATUS"); statusStr != "" {
		status := types.CategoryStatus(statusStr)
		if !status.IsValid() {
			log.Fatalf("Invalid SEARCH_STATUS: %q (valid: OPEN, RESOLVED)", statusStr)
		}
		if opts.Filter == nil {
			opts.Filter = &types.SearchFilterInput{}
		}
		opts.Filter.Status = &status
	}

	// SEARCH_TAGS: filter by tag IDs, comma-separated (optional, e.g. "2,15")
	if tagsStr := os.Getenv("SEARCH_TAGS"); tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
		if opts.Filter == nil {
			opts.Filter = &types.SearchFilterInput{}
		}
		opts.Filter.Tags = tags
	}

	// SEARCH_FIRST: pagination size (optional, default no limit)
	if firstStr := os.Getenv("SEARCH_FIRST"); firstStr != "" {
		first, err := strconv.Atoi(firstStr)
		if err != nil {
			log.Fatalf("Invalid SEARCH_FIRST: %q (must be integer)", firstStr)
		}
		if opts.Pagination == nil {
			opts.Pagination = &types.ForwardPaginationInput{}
		}
		opts.Pagination.First = &first
	}

	// SEARCH_AFTER: pagination cursor (optional)
	if after := os.Getenv("SEARCH_AFTER"); after != "" {
		if opts.Pagination == nil {
			opts.Pagination = &types.ForwardPaginationInput{}
		}
		opts.Pagination.After = &after
	}

	// Print search parameters
	fmt.Println("=== Search Parameters ===")
	fmt.Printf("  Query:  %q\n", opts.Query)
	if opts.Filter != nil {
		if opts.Filter.Status != nil {
			fmt.Printf("  Status: %s\n", *opts.Filter.Status)
		}
		if len(opts.Filter.Tags) > 0 {
			fmt.Printf("  Tags:   %s\n", strings.Join(opts.Filter.Tags, ", "))
		}
	}
	if opts.Pagination != nil {
		if opts.Pagination.First != nil {
			fmt.Printf("  First:  %d\n", *opts.Pagination.First)
		}
		if opts.Pagination.After != nil {
			fmt.Printf("  After:  %s\n", *opts.Pagination.After)
		}
	}
	fmt.Println()

	// Execute search
	result, err := client.Search(opts)
	if err != nil {
		log.Fatalf("Error searching: %v", err)
	}

	// Print categories
	fmt.Printf("--- Categories (%d) ---\n", len(result.Categories.Edges))
	for i, edge := range result.Categories.Edges {
		cat := edge.Node
		tags := make([]string, 0, len(cat.Tags.Edges))
		for _, tagEdge := range cat.Tags.Edges {
			tags = append(tags, tagEdge.Node.Name)
		}
		fmt.Printf("[%d] %s\n", i+1, cat.Title)
		fmt.Printf("    ID:     %s\n", cat.ID)
		fmt.Printf("    Volume: $%.2f\n", cat.Statistics.VolumeTotalUsd)
		if cat.EndsAt != nil {
			fmt.Printf("    Ends:   %s\n", *cat.EndsAt)
		}
		if len(tags) > 0 {
			fmt.Printf("    Tags:   %s\n", strings.Join(tags, ", "))
		}
		fmt.Println()
	}

	// Print markets
	fmt.Printf("--- Markets (%d) ---\n", len(result.Markets.Edges))
	for i, edge := range result.Markets.Edges {
		mkt := edge.Node
		outcomes := make([]string, 0, len(mkt.Outcomes.Edges))
		for _, outcomeEdge := range mkt.Outcomes.Edges {
			outcomes = append(outcomes, outcomeEdge.Node.Name)
		}
		tags := make([]string, 0, len(mkt.Category.Tags.Edges))
		for _, tagEdge := range mkt.Category.Tags.Edges {
			tags = append(tags, tagEdge.Node.Name)
		}
		fmt.Printf("[%d] %s\n", i+1, mkt.Question)
		fmt.Printf("    ID:        %s\n", mkt.ID)
		fmt.Printf("    Volume:    $%.2f\n", mkt.Statistics.VolumeTotalUsd)
		fmt.Printf("    Precision: %d\n", mkt.DecimalPrecision)
		fmt.Printf("    Outcomes:  %s\n", strings.Join(outcomes, ", "))
		if len(tags) > 0 {
			fmt.Printf("    Tags:      %s\n", strings.Join(tags, ", "))
		}
		fmt.Printf("    Cursor:    %s\n", edge.Cursor)
		fmt.Println()
	}

	fmt.Println("==============================")
}
