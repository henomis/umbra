// content.go
package content

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
)

// Meta represents provider-specific metadata for a chunk copy.
type Meta = json.RawMessage

// Content represents a file fragmented into ordered chunks.
type Content struct {
	Hash   [32]byte `json:"hash"`   // FileHash stores the overall file hash.
	Size   int64    `json:"size"`   // Size holds the original file size in bytes.
	Chunks []Chunk  `json:"chunks"` // Chunks holds the chunk sequence.
}

// Chunk represents a single chunk of the file.
type Chunk struct {
	ID     uint32      `json:"id"`
	Hash   [32]byte    `json:"hash"`
	Size   int64       `json:"size"`
	Copies []ChunkCopy `json:"copies"`
}

// ChunkCopy represents a redundant copy of a chunk stored by a provider.
type ChunkCopy struct {
	Provider string `json:"provider"`
	Meta     Meta   `json:"meta"`
}

// New returns a new Content holding the provided file hash.
func New(fileHash [32]byte, size int64) *Content {
	return &Content{
		Hash:   fileHash,
		Size:   size,
		Chunks: make([]Chunk, 0),
	}
}

// Add stores chunk data and metadata, optionally creating a new chunk ID.
// If chunkID is nil, the method assigns the next incremental ID.
func (c *Content) Add(chunkHash [32]byte, size int64, provider string, chunkID *uint32, meta Meta) uint32 {
	id := c.nextChunkID()
	if chunkID != nil {
		id = *chunkID
	}

	if idx := c.chunkIndex(id); idx >= 0 {
		c.Chunks[idx].Copies = append(c.Chunks[idx].Copies, ChunkCopy{Provider: provider, Meta: meta})
		return id
	}

	c.appendChunk(id, chunkHash, size, provider, meta)
	return id
}

// Encode marshals Content into JSON.
func (c *Content) Encode() ([]byte, error) {
	return json.Marshal(c)
}

// NewFromData unmarshals Content from JSON data.
func NewFromData(data []byte) (*Content, error) {
	var c Content
	err := json.Unmarshal(data, &c)
	return &c, err
}

// ComputeFileHash derives the file hash by concatenating chunk hashes in order.
func (c *Content) ComputeFileHash() []byte {
	h := sha256.New()
	for _, chunk := range c.Chunks {
		h.Write(chunk.Hash[:])
	}
	return h.Sum(nil)
}

// VerifyFileHash returns true when the stored hash matches the computed one.
func (c *Content) VerifyFileHash() bool {
	return bytes.Equal(c.ComputeFileHash(), c.Hash[:])
}

func (c *Content) nextChunkID() uint32 {
	if len(c.Chunks) == 0 {
		return 1
	}
	return c.Chunks[len(c.Chunks)-1].ID + 1
}

func (c *Content) chunkIndex(id uint32) int {
	for i := range c.Chunks {
		if c.Chunks[i].ID == id {
			return i
		}
	}
	return -1
}

func (c *Content) appendChunk(id uint32, chunkHash [32]byte, size int64, provider string, meta Meta) {
	c.Chunks = append(c.Chunks, Chunk{
		ID:   id,
		Hash: chunkHash,
		Size: size,
		Copies: []ChunkCopy{
			{Provider: provider, Meta: meta},
		},
	})
}
