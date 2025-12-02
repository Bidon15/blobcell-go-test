# BlobCell Example (Go-Square v3)

A simple Go example demonstrating how to submit blobs to Celestia using go-square v3 with TOML configuration.

## Features

- Uses `go-square v3` instead of v2
- Configuration via `config.toml` (no environment variables needed)
- In-memory keyring for demo simplicity
- Submits 3 example blobs to Celestia network

## Prerequisites

- Go 1.24 or later
- Access to Celestia RPC and gRPC endpoints (local node or public endpoint)
- Private key from Keplr wallet

## Getting Your Private Key from Keplr

1. Open Keplr extension → Click account icon (top right)
2. Select "Show Private Key"
3. Enter password and copy the hex key
4. Use this key in the `config.toml` file

## Configuration

Create a `config.toml` file in the project root:

```toml
[celestia]
# Get your private key from Keplr: Settings > Show Private Key
private_key = "your_hex_private_key_here"
network = "mocha-4"
rpc_url = "http://localhost:26658"
grpc_url = "http://localhost:9090"
auth_token = ""
```

You can also copy from the example:
```bash
cp config.toml.example config.toml
# Then edit config.toml with your values
```

## Installation

```bash
go mod tidy
go build
```

## Usage

```bash
# Make sure config.toml exists with your settings
./blobcell-example
```

## What It Does

The example:
1. Loads configuration from `config.toml`
2. Creates an in-memory keyring (no files stored locally)
3. Connects to the Celestia network
4. Submits 3 test blobs with unique timestamps
5. Verifies each blob by retrieving it
6. Outputs success messages and block heights

## Output Example

```
Submitting 3 blobs to Celestia...

✓ Blob 1 submitted at height 12345
 ✓ Verified: Hello from BlobCell! Message #1 at 2025-12-01T21:00:00Z
✓ Blob 2 submitted at height 12346
 ✓ Verified: Hello from BlobCell! Message #2 at 2025-12-01T21:00:02Z
✓ Blob 3 submitted at height 12347
 ✓ Verified: Hello from BlobCell! Message #3 at 2025-12-01T21:00:04Z

🎉 All 3 blobs submitted successfully!
View your blobs on https://mocha.celenium.io
```

## Security Note

**Never commit your `config.toml` file to git!** Add it to your `.gitignore`:

```
config.toml
```

## Configuration Options

| Field | Description | Example |
|-------|-------------|---------|
| `celestia.private_key` | Your hex private key from Keplr | `"abc123..."` |
| `celestia.network` | Celestia network name | `"mocha-4"` |
| `celestia.rpc_url` | RPC endpoint URL | `"http://localhost:26658"` |
| `celestia.grpc_url` | gRPC endpoint URL | `"http://localhost:9090"` |
| `celestia.auth_token` | Optional auth token | `""` (leave empty) |

## Dependencies

- `github.com/celestiaorg/go-square/v3` - Celestia's data availability sampling library
- `github.com/celestiaorg/celestia-node` - Celestia node API client
- `github.com/cosmos/cosmos-sdk` - Cosmos SDK for keyring management
- `github.com/spf13/viper` - Configuration management

## License

This is a demonstration example for educational purposes.
