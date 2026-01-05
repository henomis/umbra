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
	"github.com/henomis/umbra/internal/provider"
)

// Upload orchestrates the chunk sizing, encryption setup, content creation, and
// manifest generation for the configured Umbra instance.
func (u *Umbra) Upload(ctx context.Context) error {
	// calculate chunk size
	chunkSize, chunks, fileSize, err := u.calculateChunkSize()
	if err != nil {
		return fmt.Errorf("failed to calculate chunk size: %w", err)
	}

	// check chunk size against providers' max
	if chunkSize > u.getMaxChunkSizeForProviders() {
		return ErrChunkSizeExceedsProviderLimit
	}

	// create crypto and manifest
	crypto, err := crypto.New([]byte(u.config.Password))
	if err != nil {
		return fmt.Errorf("failed to create crypto: %w", err)
	}

	content, err := u.createContent(ctx, fileSize, chunks, chunkSize, crypto)
	if err != nil {
		return fmt.Errorf("failed to create content: %w", err)
	}

	contentData, err := content.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode content: %w", err)
	}

	manifestData := bytes.NewBuffer(nil)
	manifest := manifest.New(crypto)
	if err := manifest.Encode(manifestData, contentData); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	if err := u.saveManifest(ctx, manifestData.Bytes()); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	expire := u.getProviderMinExpireDuration()

	if !u.config.Quiet {
		fmt.Printf("✅ Upload completed. Manifest '%s' expires in: %s\n", u.config.ManifestPath, expire.String())
	}

	return nil
}

// createContent builds the content manifest by reading the input file in
// chunkSize increments, hashing each chunk, and delegating encryption and
// upload to processChunk while reusing the provided crypto helper.
func (u *Umbra) createContent(ctx context.Context, size, nChunks, chunkSize int64, crypto *crypto.Crypto) (*content.Content, error) {
	buffer := make([]byte, chunkSize)

	var bar *mpb.Bar

	if !u.config.Quiet {
		bar = u.progress.New(
			nChunks*int64(u.config.Upload.Copies),
			mpb.BarStyle().Rbound("|"),
			mpb.PrependDecorators(
				decor.Name("Uploading: ", decor.WC{W: 12}),
				decor.CountersNoUnit("%d/%d", decor.WCSyncWidth),
			),
			mpb.AppendDecorators(
				decor.Percentage(),
			),
		)
	}

	// calculate file hash
	fileHash, err := fileSHA256(u.config.Upload.InputFilePath)
	if err != nil {
		return nil, err
	}

	// create content
	content := content.New(fileHash, size)

	inputFile, err := os.Open(u.config.Upload.InputFilePath)
	if err != nil {
		return nil, err
	}
	defer inputFile.Close()

	for {
		n, err := inputFile.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if n == 0 {
			break
		}

		chunkData := buffer[:n]

		err = u.createChunk(ctx, content, chunkData, crypto, bar)
		if err != nil {
			return nil, err
		}
	}

	if bar != nil {
		bar.Wait()
	}

	return content, nil
}

// createChunk encrypts the given chunk, uploads it to the configured providers,
// and records the resulting metadata into the content manifest.
func (u *Umbra) createChunk(ctx context.Context, content *content.Content, chunkData []byte, crypto *crypto.Crypto, bar *mpb.Bar) error {
	providers := make([]provider.Provider, 0)
	var chunkID *uint32

	// encrypt chunk
	chunkHash := sha256.Sum256(chunkData)
	encryptedChunkData, err := crypto.Encode(chunkData, chunkHash[:])
	if err != nil {
		return err
	}

	for range u.config.Upload.Copies {
		provider, err := u.getUniqueRadomProvider(providers)
		if err != nil {
			return err
		}

		meta, err := provider.Upload(ctx, encryptedChunkData)
		if err != nil {
			return err
		}

		providers = append(providers, provider)
		id := content.Add(chunkHash, int64(len(chunkData)), provider.Name(), chunkID, meta)
		if chunkID == nil {
			chunkID = &id
		}

		if bar != nil {
			bar.Increment()
		}
	}

	return nil
}

// calculateChunkSize determines the chunk size based on configuration and file size.
func (u *Umbra) calculateChunkSize() (int64, int64, int64, error) {
	fileInfo, err := os.Stat(u.config.Upload.InputFilePath)
	if err != nil {
		return -1, -1, -1, err
	}

	fileSize := fileInfo.Size()
	chunkSize := u.config.Upload.ChunkSize
	if u.config.Upload.Chunks > 0 {
		chunkSize = (fileSize / int64(u.config.Upload.Chunks)) + 1
	}

	chunks := (fileSize + chunkSize - 1) / chunkSize

	return chunkSize, chunks, fileSize, nil
}

// saveManifest saves the manifest data to the configured path, optionally
// encoding it using ghost mode or uploading it to a provider.
func (u *Umbra) saveManifest(ctx context.Context, data []byte) error {
	result := bytes.NewBuffer(nil)
	var err error

	ghostMode := u.config.GhostMode
	switch ghostMode {
	case ghost.Image:
		err = ghost.EncodeToImage(result, data)
	case ghost.QRCode:
		err = ghost.EncodeToQR(result, data)
	default:
		_, err = result.Write(data)
	}

	if err != nil {
		return err
	}

	if !strings.HasPrefix(u.config.ManifestPath, "provider:") {
		return os.WriteFile(u.config.ManifestPath, result.Bytes(), 0o644)
	}

	provider, err := u.getProviderByName(strings.TrimPrefix(u.config.ManifestPath, "provider:"))
	if err != nil {
		return err
	}

	meta, err := provider.Upload(ctx, result.Bytes())
	if err != nil {
		return err
	}

	u.config.ManifestPath += ":" + base64.StdEncoding.EncodeToString(meta)

	return nil
}

// fileSHA256 computes the SHA-256 hash of the file at the given path and returns
// it as a fixed-length 32-byte array, along with any error encountered during
// reading.
func fileSHA256(path string) ([32]byte, error) {
	var zero [32]byte

	f, err := os.Open(path)
	if err != nil {
		return zero, err
	}
	defer f.Close()

	h := sha256.New()

	if _, err := io.Copy(h, f); err != nil {
		return zero, err
	}

	var sum [32]byte
	copy(sum[:], h.Sum(nil))
	return sum, nil
}
