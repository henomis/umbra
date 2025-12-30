package umbra

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/henomis/umbra/internal/content"
	"github.com/henomis/umbra/internal/crypto"
	"github.com/henomis/umbra/internal/manifest"
)

// Info orchestrates the manifest reading, decryption setup, and content retrieval
// for displaying information about the stored content.
func (u *Umbra) Info(_ context.Context) error {
	manifestData, err := u.getManifestData(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get manifest data: %w", err)
	}

	// create crypto and decode manifest
	crypto, err := crypto.New([]byte(u.config.Password))
	if err != nil {
		return fmt.Errorf("failed to create crypto: %w", err)
	}

	// decode manifest
	manifest := manifest.New(crypto)
	contentData, err := manifest.Decode(bytes.NewReader(manifestData))
	if err != nil {
		return fmt.Errorf("failed to decode manifest: %w", err)
	}

	// create content from decoded data
	content, err := content.NewFromData(contentData)
	if err != nil {
		return fmt.Errorf("failed to create content from data: %w", err)
	}

	printManifest(manifest, content)

	return nil
}

func printManifest(manifest *manifest.Manifest, content *content.Content) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintf(w, "Manifest Version:\t%d\n", manifest.Version())
	fmt.Fprintf(w, "Crypto Parameters:\n")
	cryptoParams := manifest.CryptoParameters()
	fmt.Fprintf(w, "\tCipher:\t%d\n", cryptoParams.Cipher)
	fmt.Fprintf(w, "\tKDF:\t%d\n", cryptoParams.KDF)
	fmt.Fprintf(w, "\tSalt:\t%x\n", cryptoParams.Salt)
	fmt.Fprintf(w, "\tNonce:\t%x\n\n", cryptoParams.Nonce)

	fmt.Fprintf(w, "File size:\t%d bytes\n", content.Size)
	fmt.Fprintf(w, "File hash:\t%x\n", content.Hash)
	fmt.Fprintf(w, "Chunks:\t%d\n\n", len(content.Chunks))

	for i, chunk := range content.Chunks {
		fmt.Fprintf(w, "Chunk %d:\n", i)
		fmt.Fprintf(w, "\tSize:\t%d bytes\n", chunk.Size)
		fmt.Fprintf(w, "\tHash:\t%x\n", chunk.Hash)

		fmt.Fprintf(w, "\tCopies:\t%d\n", len(chunk.Copies))
		for j, copy := range chunk.Copies {
			fmt.Fprintf(w, "\t\tCopy %d:\n", j)
			fmt.Fprintf(w, "\t\t\tProvider:\t%s\n", copy.Provider)
			fmt.Fprintf(w, "\t\t\tMeta:\t%s\n", string(copy.Meta))
		}
		fmt.Fprintln(w)
	}

	w.Flush()
}
