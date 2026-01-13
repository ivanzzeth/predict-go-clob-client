# Get Auth Message

## Overview

Retrieve the signature message used to generate a JWT token.

**Important**: Do not hardcode this message, as it is dynamic and may change.

## API Information

- **Path**: `/v1/auth/message`
- **Method**: `GET`
- **Tag**: Authorization
- **Operation ID**: `getAuthMessage`

## Authentication Requirements

- **API Key**: Required
  - Header: `x-api-key`
  - Description: API Key authentication for obtaining JWT token

## Request Parameters

No request parameters.

## Response

### Success Response (200)

**Content-Type**: `application/json`

#### Response Structure

```json
{
  "success": true,
  "data": {
    "message": "string"
  }
}
```

#### Field Description

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| success | boolean | Yes | Whether the request was successful |
| data | object | Yes | Response data |
| data.message | string | Yes | Message content to be signed |

## SDK Usage Example

```go
package main

import (
    "fmt"
    "log"

    "github.com/ivanzzeth/predict-go-clob-client"
)

func main() {
    // Create client (API Key required)
    client, err := predictclob.NewClient(
        predictclob.WithAPIKey("your-api-key"),
        predictclob.WithAPIHost("https://api.predict.fun"),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Get authentication message
    authMessage, err := client.GetAuthMessage()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Message to sign: %s\n", authMessage.Message)
}
```

## Complete Authentication Flow Example

```go
package main

import (
    "fmt"
    "log"

    "github.com/ivanzzeth/predict-go-clob-client"
)

func main() {
    // Create client
    client, err := predictclob.NewClient(
        predictclob.WithAPIKey("your-api-key"),
        predictclob.WithAPIHost("https://api.predict.fun"),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Complete authentication flow (automatically: get message -> sign -> get JWT token -> set token)
    token, address, err := client.Authenticate("your-private-key-hex")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Authenticated as %s\n", address.Hex())
    fmt.Printf("JWT Token: %s\n", token)
}
```

## Error Handling

Possible error scenarios:

- **401 Unauthorized**: Invalid or missing API Key
- **Network Error**: Request failure, timeout, etc.
- **Parse Error**: Incorrect response format

## Related APIs

- [Get JWT Token from Signature](./get-jwt-token.md) - Get JWT token using signed message

## Environments

- **Test Environment**: `https://api-testnet.predict.fun`
- **Production Environment**: `https://api.predict.fun`

## OpenAPI Specification

For the complete OpenAPI specification of this endpoint, please refer to the OpenAPI documentation in the project root directory.
