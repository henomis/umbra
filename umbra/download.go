package umbra

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"

	"github.com/henomis/umbra/internal/content"
	"github.com/henomis/umbra/internal/crypto"
	"github.com/henomis/umbra/internal/ghost"
	"github.com/henomis/umbra/internal/manifest"
)

// Download orchestrates the manifest reading, decryption setup, content retrieval, and
// output file reconstruction for the configured Umbra instance.
func (u *Umbra) Download(ctx context.Context) error {
	// read manifest data
	manifestData, err := u.getManifestData(ctx)
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

	// create output file
	outputFile, err := os.Create(u.config.Download.OutputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// process content
	err = u.extractContent(ctx, content, crypto, outputFile)
	if err != nil {
		return fmt.Errorf("failed to extract content: %w", err)
	}

	outputFileHash, err := fileSHA256(u.config.Download.OutputFilePath)
	if err != nil {
		return fmt.Errorf("failed to compute output file hash: %w", err)
	}

	if outputFileHash != content.Hash {
		return ErrOutputFileHashMismatch
	}

	if !u.config.Quiet {
		fmt.Printf("✅ Download completed. Output file: '%s'\n", u.config.Download.OutputFilePath)
	}

	return nil
}

func (u *Umbra) getManifestData(ctx context.Context) ([]byte, error) {
	// read manifest data based on ghost mode
	var data io.Reader
	var manifestData []byte
	var err error

	if strings.HasPrefix(u.config.ManifestPath, "provider:") {
		data, err = u.getManifestFromProvider(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get manifest data from URL: %w", err)
		}
	} else {
		file, err := os.Open(u.config.ManifestPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open manifest file: %w", err)
		}
		defer file.Close()

		data = file
	}

	switch u.config.GhostMode {
	case ghost.Image:
		manifestData, err = ghost.DecodeFromImage(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decode manifest from image: %w", err)
		}
	case ghost.QRCode:
		manifestData, err = ghost.DecodeFromQR(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decode manifest from qrcode: %w", err)
		}
	default:
		manifestData, err = io.ReadAll(data)
		if err != nil {
			return nil, fmt.Errorf("failed to read manifest file: %w", err)
		}
	}

	return manifestData, nil
}

func (u *Umbra) getManifestFromProvider(ctx context.Context) (io.Reader, error) {
	urlParts := strings.SplitN(u.config.ManifestPath, ":", 3)
	if len(urlParts) < 2 {
		return nil, fmt.Errorf("invalid manifest %s", u.config.ManifestPath)
	}

	provider, err := u.getProviderByName(urlParts[1])
	if err != nil {
		return nil, err
	}

	meta, err := base64.StdEncoding.DecodeString(urlParts[2])
	if err != nil {
		return nil, err
	}

	data, err := provider.Download(ctx, meta)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(data), nil
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
