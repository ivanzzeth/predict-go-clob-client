package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	predictclob "github.com/ivanzzeth/predict-go-clob-client"
	"github.com/ivanzzeth/predict-go-clob-client/constants"
	"github.com/ivanzzeth/predict-go-clob-client/types"
)

func main() {
	// Load .env file from project root directory
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get API key from environment
	apiKey := os.Getenv("PREDICT_API_KEY")

	// Create client
	var client *predictclob.Client
	var err error
	if apiKey != "" {
		client, err = predictclob.NewClient(
			predictclob.WithAPIHost(constants.DefaultAPIHost),
			predictclob.WithAPIKey(apiKey),
		)
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}
	} else {
		client = predictclob.NewReadOnlyClient(constants.DefaultAPIHost)
	}

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

	categories, err := client.GetCategories(opts)
	if err != nil {
		log.Fatalf("Error getting categories: %v", err)
	}

	// Print result using %+v to show all fields
	fmt.Printf("Total categories: %d\n\n", len(categories))
	for i, category := range categories {
		fmt.Printf("Category [%d]:\n%+v\n\n", i+1, category)
	}
}
