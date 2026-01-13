# SDK Refactoring Plan

Based on the development principles in `agent/golang/sdk_dev.md`, review and refactor the current code.

## List of Principle Violations

### 🔴 Critical Issues (Must Fix)

#### 1. Prohibition of `interface{}` and `any`

**Problem Locations:**
- `types/types.go:16` - `APIBaseResponse.Data interface{}`
- `types/types.go:43-47` - Multiple `interface{}` usage in `Market.UnmarshalJSON`
- `types/types.go:151-152` - `interface{}` usage in `Orderbook.UnmarshalJSON`
- `types/types.go:262-264` - `interface{}` usage in `Sale.UnmarshalJSON`
- `types/account.go:28-29` - `interface{}` usage in `Account.UnmarshalJSON`
- `types/account.go:63-64` - `interface{}` usage in `Referral.UnmarshalJSON`
- `types/category.go:126` - `interface{}` usage in `Category.UnmarshalJSON`
- `types/activity.go:96` - `interface{}` usage in `Activity.UnmarshalJSON`
- `order.go:172-193` - `PlaceOrder` uses `map[string]interface{}`
- `order.go:277-278` - `CancelOrder` uses `map[string]interface{}`

**Impact:**
- Violates strong typing principle
- Loses compile-time type checking
- Increases runtime error risk

**Refactoring Solution:**
- Create concrete type definitions for all requests/responses
- Use generics or concrete types to replace `interface{}`
- Remove all temporary `interface{}` structures in `UnmarshalJSON`

---

#### 2. Financial Fields Not Using `decimal.Decimal`

**Problem Locations:**
- `types/types.go:117-121` - All price/quantity fields in `MarketStats` are `string`
  - `Volume string`
  - `OpenInterest string`
  - `BidPrice string`
  - `AskPrice string`
  - `LastPrice string`
- `types/types.go:127-128` - Price and amount in `OrderbookLevel` are `string`
  - `Price string`
  - `Amount string`
- `types/types.go:251-252` - Price and amount in `Sale` are `string`
  - `Price string`
  - `Amount string`

**Impact:**
- Violates the principle that financial fields must use `decimal.Decimal`
- Cannot perform precise financial calculations
- Prone to precision issues

**Refactoring Solution:**
- Add `decimal.Decimal` type for all financial fields
- Keep `RawXXX` fields to store original string values
- Implement `UnmarshalJSON` to convert wei to decimal

---

#### 3. Request Body Using `map[string]interface{}`

**Problem Locations:**
- `order.go:172` - `PlaceOrder` request body
- `order.go:277` - `CancelOrder` request body

**Impact:**
- Violates strong typing principle
- Cannot check field correctness at compile time
- Prone to runtime errors

**Refactoring Solution:**
- Create `PlaceOrderRequest` and `CancelOrderRequest` types
- Use structs to replace `map[string]interface{}`

---

### 🟡 Medium Issues (Recommended to Fix)

#### 4. Inconsistent Time Types

**Problem Locations:**
- `types/types.go:314` - `UnixTimestamp` type
- `types/order.go:29` - `UnixTime` type

**Impact:**
- Two similar time types exist, which may cause confusion
- Does not comply with "unified time type" principle

**Refactoring Solution:**
- Unify to use one time type (recommend using `UnixTime`)
- Ensure all time fields use the unified type
- Remove `UnixTimestamp` or merge into `UnixTime`

---

#### 5. Similar Types Not Abstracted

**Problem Locations:**
- `Market` and `CategoryMarket` have many similar fields
- Multiple types have similar `UnmarshalJSON` implementations

**Impact:**
- Code duplication
- High maintenance cost
- Violates "similar types must abstract common parts" principle

**Refactoring Solution:**
- Create `BaseMarket` struct containing common fields
- `Market` and `CategoryMarket` embed `BaseMarket`
- Unify `UnmarshalJSON` implementation logic

---

#### 6. Incomplete API Method Comments

**Problem Locations:**
- Some API methods lack detailed comment descriptions
- Missing parameter descriptions and return value descriptions
- Missing error scenario descriptions

**Impact:**
- Does not comply with "must comment structs/methods based on API documentation" principle
- Reduces code readability and maintainability

**Refactoring Solution:**
- Add complete GoDoc comments for all API methods
- Include method description, parameter description, return value description, error scenario description
- Reference OpenAPI documentation to supplement detailed information

---

### 🟢 Minor Issues (Optional Optimization)

#### 7. `OrderbookLevel` Price and Amount Should Use decimal

**Problem Locations:**
- `types/types.go:126-129` - `OrderbookLevel`

**Impact:**
- Although `string` type already exists, it should be converted to `decimal.Decimal` for calculations

**Refactoring Solution:**
- Add `Price decimal.Decimal` and `Amount decimal.Decimal`
- Keep `RawPrice` and `RawAmount` fields

---

#### 8. Example Code Organization

**Current Status:**
- Example code is already organized by module, but can be further optimized

**Suggestions:**
- Ensure all examples clearly print all fields
- Add more edge case examples

---

## Refactoring Priority

### Phase 1: Core Type Refactoring (High Priority)
1. ✅ Remove all `interface{}` usage
2. ✅ Convert financial fields to `decimal.Decimal`
3. ✅ Create request body types to replace `map[string]interface{}`

### Phase 2: Type Unification and Abstraction (Medium Priority)
4. ✅ Unify time types
5. ✅ Abstract similar types
6. ✅ Complete API method comments

### Phase 3: Optimization and Enhancement (Low Priority)
7. ✅ Use decimal for `OrderbookLevel`
8. ✅ Optimize example code

---

## Detailed Refactoring Steps

### Step 1: Remove `interface{}` - `APIBaseResponse`

**Current Code:**
```go
type APIBaseResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data"`
    Message string      `json:"message,omitempty"`
}
```

**After Refactoring:**
- Remove `APIBaseResponse`, each API directly defines its own response type
- Or use generics: `type APIBaseResponse[T any] struct { Data T }`

---

### Step 2: Remove `interface{}` - `UnmarshalJSON` Methods

**Strategy:**
- All types should directly use strongly typed fields
- Leverage `UnmarshalJSON` of types like `IntegerOrString`, `common.Address`, `common.Hash` for automatic handling
- Remove all temporary `interface{}` helper structures

**Example - Market.UnmarshalJSON:**
```go
// Current: Use interface{} temporary structure
// After refactoring: Directly use strong types, rely on UnmarshalJSON of each field type
```

---

### Step 3: Convert Financial Fields to decimal

**MarketStats Refactoring:**
```go
type MarketStats struct {
    Volume       decimal.Decimal `json:"-"`
    OpenInterest decimal.Decimal `json:"-"`
    BidPrice     decimal.Decimal `json:"-"`
    AskPrice     decimal.Decimal `json:"-"`
    LastPrice    decimal.Decimal `json:"-"`

    RawVolume       string `json:"volume"`
    RawOpenInterest string `json:"openInterest"`
    RawBidPrice     string `json:"bidPrice"`
    RawAskPrice     string `json:"askPrice"`
    RawLastPrice    string `json:"lastPrice"`

    TraderCount int `json:"traderCount,omitempty"`
}
```

**OrderbookLevel Refactoring:**
```go
type OrderbookLevel struct {
    Price      decimal.Decimal `json:"-"`
    Amount     decimal.Decimal `json:"-"`
    RawPrice   string          `json:"price"`
    RawAmount  string          `json:"amount"`
}
```

**Sale Refactoring:**
```go
type Sale struct {
    Price      decimal.Decimal `json:"-"`
    Amount     decimal.Decimal `json:"-"`
    RawPrice   string          `json:"price"`
    RawAmount  string          `json:"amount"`
    // ... other fields
}
```

---

### Step 4: Create Request Body Types

**PlaceOrderRequest:**
```go
type PlaceOrderRequest struct {
    Data struct {
        Order struct {
            Salt          string `json:"salt"`
            Maker         string `json:"maker"`
            Signer        string `json:"signer"`
            Taker         string `json:"taker"`
            TokenID       string `json:"tokenId"`
            MakerAmount   string `json:"makerAmount"`
            TakerAmount   string `json:"takerAmount"`
            Expiration    string `json:"expiration"`
            Nonce         string `json:"nonce"`
            FeeRateBps    string `json:"feeRateBps"`
            Side          int64  `json:"side"`
            SignatureType int64  `json:"signatureType"`
            Signature     string `json:"signature"`
        } `json:"order"`
        PricePerShare string `json:"pricePerShare"`
        Strategy      string `json:"strategy"`
        SlippageBps   int    `json:"slippageBps"`
    } `json:"data"`
}
```

---

### Step 5: Unify Time Types

**Decision:**
- Keep `UnixTime` (in `types/order.go`)
- Remove `UnixTimestamp` (in `types/types.go`)
- All time fields uniformly use `UnixTime`

---

### Step 6: Abstract Similar Types

**BaseMarket:**
```go
type BaseMarket struct {
    ID             MarketID     `json:"id"`
    Title          string       `json:"title"`
    Question       string       `json:"question"`
    Description    string       `json:"description"`
    Status         MarketStatus `json:"status"`
    IsNegRisk      bool         `json:"isNegRisk"`
    IsYieldBearing bool         `json:"isYieldBearing"`
    FeeRateBps     FeeRateBps   `json:"feeRateBps"`
    CreatedAt      time.Time   `json:"createdAt"`
}

type Market struct {
    BaseMarket
    TokenID        TokenID   `json:"tokenId,omitempty"`
    OutcomeTokenID TokenID  `json:"outcomeTokenId,omitempty"`
    Outcomes       []Outcome `json:"outcomes,omitempty"`
    UpdatedAt      time.Time `json:"updatedAt,omitempty"`
}

type CategoryMarket struct {
    BaseMarket
    ImageURL               string                `json:"imageUrl"`
    Resolution             *CategoryMarketResolution `json:"resolution,omitempty"`
    // ... other CategoryMarket specific fields
}
```

---

## Testing Plan

1. **Unit Tests:**
   - Add unit tests for all refactored types
   - Test `UnmarshalJSON` methods
   - Test wei to decimal conversion

2. **Integration Tests:**
   - Run all example code
   - Verify API calls work correctly
   - Verify data parsing is correct

3. **Regression Tests:**
   - Ensure existing functionality is not affected
   - Verify all fields can be parsed correctly

---

## Estimated Effort

- **Phase 1 (Core Refactoring):** 2-3 days
- **Phase 2 (Type Unification):** 1-2 days
- **Phase 3 (Optimization and Enhancement):** 1 day

**Total:** 4-6 days

---

## Notes

1. **Backward Compatibility:**
   - Refactoring may affect existing code
   - Need to update all code using these types

2. **Performance Considerations:**
   - `decimal.Decimal` has slightly lower performance than `string`, but higher precision
   - For financial applications, precision takes priority over performance

3. **Test Coverage:**
   - Must ensure 100% test coverage after refactoring
   - Especially `UnmarshalJSON` methods
