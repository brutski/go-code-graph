package embeddings

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
)

// MockClient implements EmbeddingClient for testing purposes
type MockClient struct {
	model      string
	dimensions int
}

// NewMockClient creates a new mock embedding client for testing
func NewMockClient(config Config) *MockClient {
	dimensions := 384 // Default dimensions for testing
	if dim, ok := config.Options["dimensions"].(int); ok {
		dimensions = dim
	}

	model := config.Model
	if model == "" {
		model = "mock-model"
	}

	return &MockClient{
		model:      model,
		dimensions: dimensions,
	}
}

// CreateEmbedding generates a deterministic mock embedding for the given text
func (m *MockClient) CreateEmbedding(text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("empty text provided for embedding")
	}

	// Create a deterministic seed based on the text content
	hash := sha256.Sum256([]byte(text))
	// Safe conversion to prevent integer overflow
	hashValue := binary.BigEndian.Uint64(hash[:8])
	// #nosec G115 - Integer overflow is acceptable in mock/testing context
	seed := int64(hashValue & 0x7FFFFFFFFFFFFFFF) // Ensure positive value

	// Use the seed to generate consistent embeddings for the same input
	// #nosec G404 - Using math/rand is appropriate for mock/testing purposes
	r := rand.New(rand.NewSource(seed))

	embedding := make([]float32, m.dimensions)

	// Generate embedding values with a mixture of characteristics:
	// - Some dimensions based on text length
	// - Some dimensions based on character frequencies
	// - Some random but deterministic values

	textLen := len(text)

	for i := 0; i < m.dimensions; i++ {
		var value float64

		switch i % 4 {
		case 0:
			// Length-based component
			value = float64(textLen%100)/100.0*2.0 - 1.0
		case 1:
			// Character frequency component (vowels)
			vowels := 0
			for _, ch := range text {
				if ch == 'a' || ch == 'e' || ch == 'i' || ch == 'o' || ch == 'u' ||
					ch == 'A' || ch == 'E' || ch == 'I' || ch == 'O' || ch == 'U' {
					vowels++
				}
			}
			value = float64(vowels%50)/50.0*2.0 - 1.0
		case 2:
			// Hash-based deterministic random component
			hashIndex := i / 4
			if hashIndex < len(hash) {
				value = float64(hash[hashIndex])/255.0*2.0 - 1.0
			} else {
				value = r.Float64()*2.0 - 1.0
			}
		case 3:
			// Trigonometric component for smoothness
			value = math.Sin(float64(i)*0.1 + float64(seed%1000)*0.01)
		}

		embedding[i] = float32(value)
	}

	// Normalize the embedding vector (optional, but makes it more realistic)
	m.normalizeVector(embedding)

	return embedding, nil
}

// normalizeVector normalizes the embedding vector to unit length
func (m *MockClient) normalizeVector(vec []float32) {
	var norm float64
	for _, v := range vec {
		norm += float64(v * v)
	}

	if norm == 0 {
		return
	}

	norm = math.Sqrt(norm)
	for i := range vec {
		vec[i] = float32(float64(vec[i]) / norm)
	}
}

// GetModelName returns the model name/identifier being used
func (m *MockClient) GetModelName() string {
	return m.model
}

// GetDimensions returns the embedding dimensions for this model
func (m *MockClient) GetDimensions() int {
	return m.dimensions
}

// SetDimensions allows changing the dimensions (useful for testing)
func (m *MockClient) SetDimensions(dims int) {
	m.dimensions = dims
}

// BatchCreateEmbeddings generates embeddings for multiple texts (for testing batch operations)
func (m *MockClient) BatchCreateEmbeddings(texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))

	for i, text := range texts {
		embedding, err := m.CreateEmbedding(text)
		if err != nil {
			return nil, fmt.Errorf("failed to create embedding for text %d: %w", i, err)
		}
		embeddings[i] = embedding
	}

	return embeddings, nil
}
