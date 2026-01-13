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

func main() {
	// Load .env file from project root directory
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Get API key from environment (required)
	apiKey := os.Getenv("PREDICT_API_KEY")

	// Create read-only client with API key
	client := predictclob.NewReadOnlyClient(constants.DefaultAPIHost, apiKey)

	// Example 1: Get all categories
	fmt.Println("=== Example 1: Get all categories ===")
	allCategoriesResp, err := client.GetCategories(nil)
	if err != nil {
		log.Fatalf("Failed to get all categories: %v", err)
	}
	fmt.Printf("Found %d categories\n", len(allCategoriesResp.Data))
	if allCategoriesResp.Cursor != nil {
		fmt.Printf("Next page cursor: %s\n", *allCategoriesResp.Cursor)
	}
	printCategories(allCategoriesResp.Data)

	// Example 2: Get categories by status (OPEN)
	fmt.Println("\n=== Example 2: Get OPEN categories ===")
	openOpts := &types.GetCategoriesOptions{
		Status: types.CategoryStatusOpen,
	}
	openCategoriesResp, err := client.GetCategories(openOpts)
	if err != nil {
		log.Fatalf("Failed to get OPEN categories: %v", err)
	}
	fmt.Printf("Found %d OPEN categories\n", len(openCategoriesResp.Data))
	printCategories(openCategoriesResp.Data)

	// Example 3: Get categories with pagination
	fmt.Println("\n=== Example 3: Get categories with pagination (first 5) ===")
	first := 5
	paginationOpts := &types.GetCategoriesOptions{
		First: &first,
	}
	paginatedResp, err := client.GetCategories(paginationOpts)
	if err != nil {
		log.Fatalf("Failed to get paginated categories: %v", err)
	}
	fmt.Printf("Found %d categories (requested first %d)\n", len(paginatedResp.Data), first)
	if paginatedResp.Cursor != nil {
		fmt.Printf("Next page cursor: %s\n", *paginatedResp.Cursor)
	}
	printCategories(paginatedResp.Data)

	// Example 4: Get categories with sort order
	fmt.Println("\n=== Example 4: Get categories sorted by volume (24h desc) ===")
	sortOpts := &types.GetCategoriesOptions{
		Status: types.CategoryStatusOpen,
		Sort:   types.CategorySortVolume24HDesc,
	}
	sortedResp, err := client.GetCategories(sortOpts)
	if err != nil {
		log.Fatalf("Failed to get sorted categories: %v", err)
	}
	fmt.Printf("Found %d categories (sorted by volume 24h desc)\n", len(sortedResp.Data))
	printCategories(sortedResp.Data)

	// Example 5: Get next page if cursor exists
	if paginatedResp.Cursor != nil && *paginatedResp.Cursor != "" {
		fmt.Println("\n=== Example 5: Get next page using cursor ===")
		nextOpts := &types.GetCategoriesOptions{
			First: &first,
			After: paginatedResp.Cursor,
		}
		nextResp, err := client.GetCategories(nextOpts)
		if err != nil {
			log.Fatalf("Failed to get next page: %v", err)
		}
		fmt.Printf("Found %d more categories\n", len(nextResp.Data))
		printCategories(nextResp.Data)
	}

	// Example 6: Get categories with custom filters from command line or environment
	fmt.Println("\n=== Example 6: Get categories with custom filters ===")
	opts := buildOptionsFromArgs()
	if opts != nil {
		customResp, err := client.GetCategories(opts)
		if err != nil {
			log.Fatalf("Failed to get categories with custom filters: %v", err)
		}
		fmt.Printf("Found %d categories with custom filters\n", len(customResp.Data))
		if customResp.Cursor != nil {
			fmt.Printf("Next page cursor: %s\n", *customResp.Cursor)
		}
		printCategories(customResp.Data)
	} else {
		fmt.Println("No custom filters provided (use command line args or env vars)")
	}
}

// buildOptionsFromArgs builds GetCategoriesOptions from command line arguments or environment variables
func buildOptionsFromArgs() *types.GetCategoriesOptions {
	opts := &types.GetCategoriesOptions{}

	// Status from args or env
	if len(os.Args) > 1 {
		opts.Status = types.CategoryStatus(os.Args[1])
	}
	if opts.Status == "" {
		if statusStr := os.Getenv("CATEGORY_STATUS"); statusStr != "" {
			opts.Status = types.CategoryStatus(statusStr)
		}
	}

	// First (limit) from args or env
	if len(os.Args) > 2 {
		if first, err := strconv.Atoi(os.Args[2]); err == nil {
			opts.First = &first
		}
	}
	if opts.First == nil {
		if firstStr := os.Getenv("CATEGORY_FIRST"); firstStr != "" {
			if first, err := strconv.Atoi(firstStr); err == nil {
				opts.First = &first
			}
		}
	}

	// After (cursor) from args or env
	if len(os.Args) > 3 {
		opts.After = &os.Args[3]
	}
	if opts.After == nil {
		if afterStr := os.Getenv("CATEGORY_AFTER"); afterStr != "" {
			opts.After = &afterStr
		}
	}

	// Sort from env
	if sortStr := os.Getenv("CATEGORY_SORT"); sortStr != "" {
		opts.Sort = types.CategorySort(sortStr)
	}

	// Return nil if no options are set
	if opts.Status == "" && opts.First == nil && opts.After == nil && opts.Sort == "" {
		return nil
	}

	return opts
}

// printCategories prints category details with all fields clearly displayed
func printCategories(categories []types.Category) {
	for i, category := range categories {
		fmt.Printf("\n--- Category %d ---\n", i+1)
		fmt.Printf("ID: %s\n", category.ID.String())
		fmt.Printf("Slug: %s\n", category.Slug)
		fmt.Printf("Title: %s\n", category.Title)
		fmt.Printf("Description: %s\n", category.Description)
		fmt.Printf("Image URL: %s\n", category.ImageURL)
		fmt.Printf("Is Neg Risk: %v\n", category.IsNegRisk)
		fmt.Printf("Is Yield Bearing: %v\n", category.IsYieldBearing)
		fmt.Printf("Market Variant: %s\n", category.MarketVariant.String())
		fmt.Printf("Status: %s\n", category.Status.String())
		fmt.Printf("Created At: %s\n", category.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Published At: %s\n", category.PublishedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Starts At: %s\n", category.StartsAt.Format("2006-01-02 15:04:05"))
		if category.EndsAt != nil {
			fmt.Printf("Ends At: %s\n", category.EndsAt.Format("2006-01-02 15:04:05"))
		} else {
			fmt.Printf("Ends At: (null)\n")
		}

		// Print tags
		fmt.Printf("Tags (%d):\n", len(category.Tags))
		for j, tag := range category.Tags {
			fmt.Printf("  Tag %d: ID=%s, Name=%s\n", j+1, tag.ID.String(), tag.Name)
		}

		// Print markets
		fmt.Printf("Markets (%d):\n", len(category.Markets))
		for j, market := range category.Markets {
			fmt.Printf("  Market %d:\n", j+1)
			fmt.Printf("    ID: %s\n", market.ID.String())
			fmt.Printf("    Title: %s\n", market.Title)
			fmt.Printf("    Question: %s\n", market.Question)
			fmt.Printf("    Status: %s\n", market.Status.String())
			fmt.Printf("    Is Neg Risk: %v\n", market.IsNegRisk)
			fmt.Printf("    Is Yield Bearing: %v\n", market.IsYieldBearing)
			fmt.Printf("    Fee Rate Bps: %d\n", market.FeeRateBps)
			fmt.Printf("    Category Slug: %s\n", market.CategorySlug)
			fmt.Printf("    Created At: %s\n", market.CreatedAt.Format("2006-01-02 15:04:05"))
			if market.Resolution != nil {
				fmt.Printf("    Resolution: Name=%s, IndexSet=%d, OnChainID=%s\n",
					market.Resolution.Name, market.Resolution.IndexSet, market.Resolution.OnChainID)
				if market.Resolution.Status != nil {
					fmt.Printf("      Status: %s\n", *market.Resolution.Status)
				}
			}
			fmt.Printf("    Outcomes (%d):\n", len(market.Outcomes))
			for k, outcome := range market.Outcomes {
				fmt.Printf("      Outcome %d: Name=%s, IndexSet=%d, OnChainID=%s\n",
					k+1, outcome.Name, outcome.IndexSet, outcome.OnChainID)
				if outcome.Status != nil {
					fmt.Printf("        Status: %s\n", *outcome.Status)
				}
			}
		}
	}
}
