package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	predictclob "github.com/ivanzzeth/predict-go-clob-client"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/ivanzzeth/predict-go-clob-client/types"
	"github.com/joho/godotenv"
)

// Environment variables:
// - PREDICT_API_KEY: API key (required)
// - CATEGORY_FIRST: limit number of results (optional)
// - CATEGORY_AFTER: pagination cursor (optional)
// - CATEGORY_STATUS: filter by status - OPEN or RESOLVED (optional)
// - CATEGORY_SORT: sort order (optional)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	apiKey := os.Getenv("PREDICT_API_KEY")
	client := predictclob.NewReadOnlyClient(constants.DefaultAPIHost, apiKey)

	opts := &types.GetCategoriesOptions{}

	if firstStr := os.Getenv("CATEGORY_FIRST"); firstStr != "" {
		if first, err := strconv.Atoi(firstStr); err == nil {
			opts.First = &first
		}
	}
	if after := os.Getenv("CATEGORY_AFTER"); after != "" {
		opts.After = &after
	}
	if status := os.Getenv("CATEGORY_STATUS"); status != "" {
		opts.Status = types.CategoryStatus(status)
	}
	if sort := os.Getenv("CATEGORY_SORT"); sort != "" {
		opts.Sort = types.CategorySort(sort)
	}

	resp, err := client.GetCategories(opts)
	if err != nil {
		log.Fatalf("Failed to get categories: %v", err)
	}

	fmt.Printf("Found %d categories\n", len(resp.Data))
	if resp.Cursor != nil {
		fmt.Printf("Next cursor: %s\n", *resp.Cursor)
	}

	for i, c := range resp.Data {
		fmt.Printf("\n--- Category %d ---\n", i+1)
		fmt.Printf("ID: %s\n", c.ID.String())
		fmt.Printf("Slug: %s\n", c.Slug)
		fmt.Printf("Title: %s\n", c.Title)
		fmt.Printf("Status: %s\n", c.Status.String())
		fmt.Printf("Is Yield Bearing: %v\n", c.IsYieldBearing)
		fmt.Printf("Markets (%d):\n", len(c.Markets))
		for j, m := range c.Markets {
			fmt.Printf("  %d. ID=%s, Title=%s, Status=%s\n", j+1, m.ID.String(), m.Title, m.Status.String())
			for k, o := range m.Outcomes {
				fmt.Printf("     Outcome %d: %s (OnChainID: %s)\n", k+1, o.Name, o.OnChainID)
			}
		}
	}
}
