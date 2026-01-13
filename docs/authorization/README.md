# Authorization

Authentication related API endpoints for the Predict.fun platform.

## Available APIs

- [Get Auth Message](./get-auth-message.md) - Get authentication message for signing
- [Get JWT Token from Signature](./get-jwt-token.md) - Get JWT Token using signature (to be added)

## Authentication Flow

1. Get authentication message using API Key
2. Sign the message with your Ethereum private key
3. Submit the signature to get JWT token
4. Use JWT token for authenticated API calls

For a complete example, see [Get Auth Message](./get-auth-message.md).
