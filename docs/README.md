# API Documentation

This documentation directory contains complete documentation for the Predict.fun API, organized by functional modules.

## Directory Structure

```
docs/
├── README.md                    # This document (index)
└── authorization/               # Authentication related APIs
    ├── get-auth-message.md      # Get authentication message
    └── get-jwt-token.md         # Get JWT Token (to be added)
```

## Module List

### Authorization

Authentication related API endpoints.

- [Get Auth Message](./authorization/get-auth-message.md) - Get authentication message for signing
- [Get JWT Token from Signature](./authorization/get-jwt-token.md) - Get JWT Token using signature (to be added)

## Usage Instructions

Each document includes:
- API overview and description
- Request parameters and response structure
- SDK usage examples
- Error handling instructions
- Related API links

## Quick Start

Check the example code in the [examples](../examples/) directory to learn how to use the SDK to call these APIs.

## Environment Information

- **Test Environment**: `https://api-testnet.predict.fun`
- **Production Environment**: `https://api.predict.fun`

## Related Resources

- [SDK Source Code](../)
- [Example Code](../examples/)
- [README](../README.md)
