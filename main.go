package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	popsigner "github.com/Bidon15/popsigner/sdk-go"
	"github.com/celestiaorg/celestia-node/api/client"
	"github.com/celestiaorg/celestia-node/blob"
	"github.com/celestiaorg/celestia-node/nodebuilder/p2p"
	libshare "github.com/celestiaorg/go-square/v3/share"
	"github.com/spf13/viper"
)

func main() {
	// Load config.toml
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("Failed to read config file: %v", err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Get PopSigner config
	apiKey := viper.GetString("celestia.popsigner_api_key")
	keyName := viper.GetString("celestia.key_name")

	if apiKey == "" || keyName == "" {
		panic("celestia.popsigner_api_key and celestia.key_name are required in config.toml")
	}

	// Get namespace from config
	namespaceStr := viper.GetString("celestia.namespace")
	if namespaceStr == "" {
		panic("celestia.namespace is required in config.toml")
	}

	// Setup PopSigner keyring using direct SDK method
	kr, err := popsigner.NewCelestiaKeyring(apiKey, keyName)
	if err != nil {
		panic(fmt.Sprintf("Failed to create PopSigner keyring for key '%s'. Ensure 'celestia.popsigner_api_key' is correct and 'celestia.key_name' ('%s') exists and is accessible: %v", keyName, keyName, err))
	}

	// Configure client using config values
	cfg := client.Config{
		ReadConfig: client.ReadConfig{
			BridgeDAAddr: viper.GetString("celestia.rpc_url"),
			EnableDATLS:  true,
		},
		SubmitConfig: client.SubmitConfig{
			DefaultKeyName: keyName,
			Network:        p2p.Network(viper.GetString("celestia.network")),
			CoreGRPCConfig: client.CoreGRPCConfig{
				Addr:       viper.GetString("celestia.grpc_url"),
				TLSEnabled: true,
				AuthToken:  viper.GetString("celestia.auth_token"),
			},
		},
	}

	// Create client
	c, err := client.New(ctx, cfg, kr)
	if err != nil {
		panic(fmt.Sprintf("Failed to create client: %v", err))
	}

	// Create namespace for your blobs
	// Convert string to bytes
	namespaceBytes := []byte(namespaceStr)

	// If it looks like hex, try to decode it
	if len(namespaceStr)%2 == 0 {
		if b, err := hex.DecodeString(namespaceStr); err == nil {
			namespaceBytes = b
		}
	}

	// Celestia V0 namespaces must be 28 bytes and start with leading zeros (for user namespaces).
	// We pad with leading zeros to ensure it meets the length requirement.
	if len(namespaceBytes) < 28 {
		padded := make([]byte, 28)
		copy(padded[28-len(namespaceBytes):], namespaceBytes)
		namespaceBytes = padded
	} else if len(namespaceBytes) > 28 {
		// Should not happen with typical usage of this example, but truncate or hash if too long?
		// User specifically asked for stacking with leading zeros, so we assume input fits.
		// Let's truncate from the left (lossy) or hash. Let's start with hashing if too long to maintain uniqueness?
		// Actually, given the error, let's just warn or fail, but for now hash to safety if too long.
		// For this specific error (len < 10 typically), padding is the fix.
		// If > 28, simply hashing to 28 bytes might violate the leading zero rule again if hash doesn't have them.
		// Safest strictly for the user's input (8 bytes) is padding.
		// If input was indeed > 28 bytes, we can't easily make it a valid user namespace (requires 18 zeros).
		// So we assume input is short ID.
		hash := sha256.Sum256(namespaceBytes)
		// We can't just take the hash, we need leading zeros.
		// We'll take the last 10 bytes of the hash and pad.
		padded := make([]byte, 28)
		copy(padded[28-10:], hash[:10])
		namespaceBytes = padded
	}

	namespace, err := libshare.NewNamespace(libshare.ShareVersionZero, namespaceBytes)
	if err != nil {
		panic(fmt.Sprintf("Failed to create namespace: %v", err))
	}

	// Submit blobs
	if err := submitBlobs(ctx, c, namespace, nil); err != nil {
		panic(err)
	}
}

func submitBlobs(ctx context.Context, c *client.Client, namespace libshare.Namespace, submitOpts *blob.SubmitOptions) error {
	// Submit 3 blobs to demonstrate the workflow
	fmt.Println("Submitting 3 blobs to Celestia...\\n")
	for i := 1; i <= 3; i++ {
		// Create unique blob data
		data := fmt.Sprintf("Hello from BlobCell! Message #%d at %s", i, time.Now().Format(time.RFC3339))

		b, err := blob.NewBlob(libshare.ShareVersionZero, namespace, []byte(data), nil)
		if err != nil {
			return fmt.Errorf("failed to create blob %d: %w", i, err)
		}

		// Submit
		height, err := c.Blob.Submit(ctx, []*blob.Blob{b}, submitOpts)
		if err != nil {
			return fmt.Errorf("failed to submit blob %d: %w", i, err)
		}

		fmt.Printf("✓ Blob %d submitted at height %d\\n", i, height)

		// Verify by retrieving
		retrieved, err := c.Blob.Get(ctx, height, namespace, b.Commitment)
		if err != nil {
			fmt.Printf(" Warning: Could not verify blob %d: %v\\n", i, err)
		} else {
			fmt.Printf(" ✓ Verified: %s\\n", string(retrieved.Data()))
		}

		// Small delay between submissions
		if i < 3 {
			time.Sleep(2 * time.Second)
		}
	}

	fmt.Println("\\n🎉 All 3 blobs submitted successfully!")
	fmt.Println("View your blobs on https://mocha.celenium.io")
	return nil
}
