# BlobCell Example (Go-Square v3)

A simple Go example demonstrating how to submit blobs to Celestia using go-square v3 with TOML configuration.

## Features

- Uses `go-square v3` instead of v2
- Configuration via `config.toml` (no environment variables needed)
- **Flexible authentication**: Supports both private keys and mnemonics
- **Custom namespace support**: Define your own namespace for blob organization
- **Feegranter support**: Optional feegranter for transaction fees
- In-memory keyring for demo simplicity (no files stored locally)
- Submits 3 example blobs to Celestia network
- Automatic blob verification after submission

## Prerequisites

- Go 1.24 or later
- Access to Celestia RPC and gRPC endpoints (local node or public endpoint)
- Either a private key from Keplr wallet OR a mnemonic phrase

## Getting Your Credentials

### Option 1: Private Key from Keplr

1. Open Keplr extension → Click account icon (top right)
2. Select "Show Private Key"
3. Enter password and copy the hex key
4. Use this key in the `config.toml` file

### Option 2: Mnemonic Phrase

Use your existing 12 or 24-word mnemonic phrase from any Cosmos wallet.

## Configuration

Create a `config.toml` file in the project root:

```toml
[celestia]
# Authentication: Choose ONE of the following
private_key = "your_hex_private_key_here"
# mnemonic = "your twelve or twenty four word mnemonic phrase here"

# Network configuration
network = "mocha-4"
rpc_url = "http://localhost:26658"
grpc_url = "http://localhost:9090"

# Namespace for your blobs (can be text or hex)
namespace = "test/blobcell"

# Optional: feegranter address for sponsored transactions
feegranter = "celestia1abc..."

# Optional: auth token for gRPC endpoint (leave empty if not needed)
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

The example demonstrates the complete workflow for working with Celestia:

1. **Configuration Loading**: Reads settings from `config.toml`
2. **Keyring Setup**: Creates an in-memory keyring from private key or mnemonic
3. **Client Creation**: Connects to Celestia network via RPC and gRPC
4. **Namespace Creation**: Generates a custom namespace for blob organization
5. **Blob Submission**: Submits 3 test blobs with unique timestamps
6. **Verification**: Retrieves each blob to confirm successful storage
7. **Output**: Displays block heights and success messages

## Output Example

```
Using feegranter: celestia1abc...
Submitting 3 blobs to Celestia...

✓ Blob 1 submitted at height 12345
 ✓ Verified: Hello from BlobCell! Message #1 at 2025-12-03T16:00:00Z
✓ Blob 2 submitted at height 12346
 ✓ Verified: Hello from BlobCell! Message #2 at 2025-12-03T16:00:02Z
✓ Blob 3 submitted at height 12347
 ✓ Verified: Hello from BlobCell! Message #3 at 2025-12-03T16:00:04Z

🎉 All 3 blobs submitted successfully!
View your blobs on https://mocha.celenium.io
```

## Security Note

**Never commit your `config.toml` file to git!** It contains sensitive credentials. The `.gitignore` file already excludes it:

```
config.toml
```

## Configuration Options

| Field | Required | Description | Example |
|-------|----------|-------------|---------|
| `celestia.private_key` | Either this or mnemonic | Your hex private key from Keplr | `"abc123..."` |
| `celestia.mnemonic` | Either this or private_key | Your 12/24 word mnemonic phrase | `"word1 word2..."` |
| `celestia.network` | Yes | Celestia network name | `"mocha-4"` |
| `celestia.rpc_url` | Yes | RPC endpoint URL | `"http://localhost:26658"` |
| `celestia.grpc_url` | Yes | gRPC endpoint URL | `"http://localhost:9090"` |
| `celestia.namespace` | Yes | Namespace for your blobs (text or hex) | `"test/blobcell"` |
| `celestia.feegranter` | No | Feegranter address for sponsored txs | `"celestia1abc..."` |
| `celestia.auth_token` | No | Auth token for gRPC endpoint | `""` (leave empty) |

## Key Concepts

### In-Memory Keyring

This demo uses an in-memory keyring that doesn't persist to disk. This is ideal for:
- Testing and development
- CI/CD pipelines
- Temporary operations
- Avoiding keystore file management

### Namespace

Namespaces organize blobs on Celestia. They can be:
- Plain text (gets converted to bytes)
- Hex strings (gets decoded)
- Automatically hashed if too long (>10 bytes for v0)

### Feegranter

If you have access to a feegranter address, you can submit blobs without paying fees directly. The feegranter pays on your behalf.

## Dependencies

- `github.com/celestiaorg/go-square/v3` - Celestia's data availability sampling library
- `github.com/celestiaorg/celestia-node` - Celestia node API client
- `github.com/cosmos/cosmos-sdk` - Cosmos SDK for keyring management
- `github.com/spf13/viper` - Configuration management

## Troubleshooting

### "Either celestia.private_key or celestia.mnemonic is required"
Make sure you've set at least one authentication method in `config.toml`.

### "Failed to create namespace"
Check that your namespace string is valid. It should be either plain text or a hex string.

### "Failed to submit blob"
Verify your RPC/gRPC endpoints are accessible and your credentials are correct.

## License

This is a demonstration example for educational purposes.
