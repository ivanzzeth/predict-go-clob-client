package constants

// Default API host
const (
	DefaultAPIHost = "https://api.predict.fun"
)

// API Endpoints
const (
	// Root endpoint
	EndpointRoot = "/"

	// Category endpoints
	EndpointCategories     = "/v1/categories"
	EndpointCategoryBySlug = "/v1/categories/%s"

	// Market endpoints
	EndpointMarkets        = "/v1/markets"
	EndpointMarketByID     = "/v1/markets/%s"
	EndpointMarketStats    = "/v1/markets/%s/stats"
	EndpointMarketOrderbook = "/v1/markets/%s/orderbook"
	EndpointMarketSale     = "/v1/markets/%s/sale"
)
