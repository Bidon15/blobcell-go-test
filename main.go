package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/celestiaorg/celestia-node/api/client"
	"github.com/celestiaorg/celestia-node/blob"
	"github.com/celestiaorg/celestia-node/nodebuilder/p2p"
	"github.com/celestiaorg/celestia-node/state"
	libshare "github.com/celestiaorg/go-square/v3/share"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
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

	// Get private key or mnemonic from config
	privateKeyHex := viper.GetString("celestia.private_key")
	mnemonic := viper.GetString("celestia.mnemonic")

	if privateKeyHex == "" && mnemonic == "" {
		panic("Either celestia.private_key or celestia.mnemonic is required in config.toml")
	}

	// Get namespace from config
	namespaceStr := viper.GetString("celestia.namespace")
	if namespaceStr == "" {
		panic("celestia.namespace is required in config.toml")
	}

	// Setup in-memory keyring
	keyname := "blobcell"
	kr, err := setupInMemoryKeyring(keyname, privateKeyHex, mnemonic)
	if err != nil {
		panic(err)
	}

	// Configure client using config values
	cfg := client.Config{
		ReadConfig: client.ReadConfig{
			BridgeDAAddr: viper.GetString("celestia.rpc_url"),
			EnableDATLS:  true,
		},
		SubmitConfig: client.SubmitConfig{
			DefaultKeyName: keyname,
			Network:        p2p.Network(viper.GetString("celestia.network")),
			CoreGRPCConfig: client.CoreGRPCConfig{
				Addr:       viper.GetString("celestia.grpc_url"),
				TLSEnabled: true,
				AuthToken:  "2ae91d0d78ef0a253990449d0bb7e9f054f024c0",
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

	// Ensure it fits in 10 bytes (v0 namespace ID size)
	if len(namespaceBytes) > 10 {
		// Hash to get 10 bytes
		hash := sha256.Sum256(namespaceBytes)
		namespaceBytes = hash[:10]
	}

	namespace, err := libshare.NewNamespace(libshare.ShareVersionZero, namespaceBytes)
	if err != nil {
		panic(fmt.Sprintf("Failed to create namespace: %v", err))
	}

	// Get feegranter from config
	feegranterStr := viper.GetString("celestia.feegranter")
	var submitOpts *blob.SubmitOptions
	if feegranterStr != "" {
		submitOpts = state.NewTxConfig(state.WithFeeGranterAddress(feegranterStr))
		fmt.Printf("Using feegranter: %s\\n", feegranterStr)
	}

	// Submit blobs
	if err := submitBlobs(ctx, c, namespace, submitOpts); err != nil {
		panic(err)
	}
}

func setupInMemoryKeyring(keyname, privateKeyHex, mnemonic string) (keyring.Keyring, error) {
	// Create minimal codec for keyring
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)
	kr := keyring.NewInMemory(cdc)

	if privateKeyHex != "" {
		// Decode hex private key
		privateKeyBytes, err := hex.DecodeString(privateKeyHex)
		if err != nil {
			return nil, fmt.Errorf("failed to decode private key hex: %w", err)
		}

		// Import as armor format (which is what Keplr exports)
		err = kr.ImportPrivKey(keyname, string(privateKeyBytes), "")
		if err != nil {
			return nil, fmt.Errorf("failed to import private key: %w", err)
		}
	} else {
		// Use mnemonic
		// Default HD path for Cosmos/Celestia: m/44'/118'/0'/0/0
		hdPath := "m/44'/118'/0'/0/0"
		// Use secp256k1
		algo := hd.Secp256k1

		_, err := kr.NewAccount(keyname, mnemonic, "", hdPath, algo)
		if err != nil {
			return nil, fmt.Errorf("failed to create account from mnemonic: %w", err)
		}
	}

	return kr, nil
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
