package constants

// Default API host
const (
	DefaultAPIHost    = "https://api.predict.fun"
	DefaultGraphQLHost = "https://graphql.predict.fun"
)

// GraphQL endpoints
const (
	EndpointGraphQL = "/graphql"
)

// Token decimals
const (
	// TokenDecimals is the number of decimals for token amounts (18 for ERC20 on BNB)
	TokenDecimals = 18
)

// API Endpoints
const (
	// Root endpoint
	EndpointRoot = "/"

	// Category endpoints
	EndpointCategories     = "/v1/categories"
	EndpointCategoryBySlug = "/v1/categories/%s"

	// Market endpoints
	EndpointMarkets         = "/v1/markets"
	EndpointMarketByID      = "/v1/markets/%s"
	EndpointMarketStats     = "/v1/markets/%s/stats"
	EndpointMarketOrderbook = "/v1/markets/%s/orderbook"
	EndpointMarketSale      = "/v1/markets/%s/sale" // Deprecated: use EndpointMarketLastSale
	EndpointMarketLastSale  = "/v1/markets/%s/last-sale"

	// Order endpoints
	EndpointOrders        = "/v1/orders"
	EndpointOrdersRemove  = "/v1/orders/remove"
	EndpointOrdersMatches = "/v1/orders/matches"
)
