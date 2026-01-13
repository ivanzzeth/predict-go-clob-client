package main

import (
	"fmt"
	"log"
	"os"

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

	// Get API key from environment (required)
	apiKey := os.Getenv("PREDICT_API_KEY")

	// Create read-only client with API key
	client := predictclob.NewReadOnlyClient(constants.DefaultAPIHost, apiKey)

	// Get status filter from command line or environment
	status := types.CategoryStatus("")
	if len(os.Args) > 1 {
		status = types.CategoryStatus(os.Args[1])
	}
	if status == "" {
		if statusStr := os.Getenv("CATEGORY_STATUS"); statusStr != "" {
			status = types.CategoryStatus(statusStr)
		}
	}

	// Call API
	var opts *types.GetCategoriesOptions
	if status != "" {
		opts = &types.GetCategoriesOptions{
			Status: status,
		}
	}

	response, err := client.GetCategories(opts)
	if err != nil {
		log.Fatalf("Error getting categories: %v", err)
	}

	// Print result using %+v to show all fields
	fmt.Printf("Total categories: %d\n", len(response.Data))
	if response.Cursor != nil {
		fmt.Printf("Next cursor: %s\n", *response.Cursor)
	}
	fmt.Println()
	for i, category := range response.Data {
		fmt.Printf("Category [%d]:\n%+v\n\n", i+1, category)
	}
}
