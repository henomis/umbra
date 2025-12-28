package umbra

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"

	"github.com/henomis/umbra/internal/content"
	"github.com/henomis/umbra/internal/crypto"
	"github.com/henomis/umbra/internal/manifest"
)

// Download orchestrates the manifest reading, decryption setup, content retrieval, and
// output file reconstruction for the configured Umbra instance.
func (u *Umbra) Download(ctx context.Context) error {
	// open manifest file
	manifestFile, err := os.Open(u.config.ManifestPath)
	if err != nil {
		return fmt.Errorf("failed to open manifest file: %w", err)
	}
	defer manifestFile.Close()

	// create crypto and decode manifest
	crypto, err := crypto.New([]byte(u.config.Password))
	if err != nil {
		return fmt.Errorf("failed to create crypto: %w", err)
	}

	// decode manifest
	manifest := manifest.New(crypto)
	contentData, err := manifest.Decode(manifestFile)
	if err != nil {
		return fmt.Errorf("failed to decode manifest: %w", err)
	}

	// create content from decoded data
	content, err := content.NewFromData(contentData)
	if err != nil {
		return fmt.Errorf("failed to create content from data: %w", err)
	}

	// create output file
	outputFile, err := os.Create(u.config.OutputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// process content
	err = u.extractContent(ctx, content, crypto, outputFile)
	if err != nil {
		return fmt.Errorf("failed to extract content: %w", err)
	}

	outputFileHash, err := fileSHA256(u.config.OutputFilePath)
	if err != nil {
		return fmt.Errorf("failed to compute output file hash: %w", err)
	}

	if outputFileHash != content.Hash {
		return ErrOutputFileHashMismatch
	}

	return nil
}

func (u *Umbra) extractContent(ctx context.Context, content *content.Content, crypto *crypto.Crypto, outputFile *os.File) error {
	var bar *mpb.Bar

	if !u.config.Quiet {
		bar = u.progress.New(
			int64(len(content.Chunks)),
			mpb.BarStyle().Rbound("|"),
			mpb.PrependDecorators(
				decor.Name("Downloading: ", decor.WC{W: 12}),
				decor.CountersNoUnit("%d/%d", decor.WCSyncWidth),
			),
			mpb.AppendDecorators(
				decor.Percentage(),
			),
		)
	}

	for _, chunk := range content.Chunks {
		err := u.extractChunk(ctx, &chunk, crypto, outputFile)
		if err != nil {
			return err
		}

		if bar != nil {
			bar.Increment()
		}
	}

	if bar != nil {
		bar.Wait()
	}

	return nil
}

func (u *Umbra) extractChunk(ctx context.Context, chunk *content.Chunk, crypto *crypto.Crypto, outputFile *os.File) error {
	var chunkErr error

	for _, c := range chunk.Copies {
		provider, err := u.getProviderByName(c.Provider)
		if err != nil {
			chunkErr = err
			continue
		}

		encryptedChunkData, err := provider.Download(ctx, c.Meta)
		if err != nil {
			chunkErr = err
			continue
		}

		chunkData, err := crypto.Decode(encryptedChunkData, chunk.Hash[:])
		if err != nil {
			chunkErr = err
			continue
		}

		chunkDataHash := sha256.Sum256(chunkData)
		if chunkDataHash != chunk.Hash {
			chunkErr = fmt.Errorf("chunk hash mismatch")
			continue
		}

		_, err = outputFile.Write(chunkData)
		if err != nil {
			chunkErr = err
			continue
		}

		// successfully processed chunk, break out of the loop
		chunkErr = nil
		break
	}

	return chunkErr
}
