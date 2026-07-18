package adapter

import (
	"context"
	"math"
	"sort"
	"sync"

	"github.com/open-strata-ai/ai-sdk-go/pkg/domain"
)

// InMemoryVectorStore is the default VectorStore adapter: in-process linear cosine search.
type InMemoryVectorStore struct {
	mu          sync.Mutex
	collections map[string][]domain.Doc
}

// NewInMemoryVectorStore builds an InMemoryVectorStore.
func NewInMemoryVectorStore() *InMemoryVectorStore {
	return &InMemoryVectorStore{collections: map[string][]domain.Doc{}}
}

func (v *InMemoryVectorStore) Upsert(ctx context.Context, collection string, docs []domain.Doc) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.collections[collection] = append(v.collections[collection], docs...)
	return nil
}

func (v *InMemoryVectorStore) Search(ctx context.Context, collection string, vec []float32, topK int) ([]domain.Hit, error) {
	v.mu.Lock()
	docs := v.collections[collection]
	cp := make([]domain.Doc, len(docs))
	copy(cp, docs)
	v.mu.Unlock()

	hits := make([]domain.Hit, 0, len(cp))
	for _, d := range cp {
		hits = append(hits, domain.Hit{ID: d.ID, Text: d.Text, Score: cosine(vec, d.Vector)})
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].Score > hits[j].Score })
	if topK > 0 && len(hits) > topK {
		hits = hits[:topK]
	}
	return hits, nil
}

func (v *InMemoryVectorStore) Delete(ctx context.Context, collection string, ids []string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	drop := map[string]bool{}
	for _, id := range ids {
		drop[id] = true
	}
	filtered := v.collections[collection][:0]
	for _, d := range v.collections[collection] {
		if !drop[d.ID] {
			filtered = append(filtered, d)
		}
	}
	v.collections[collection] = filtered
	return nil
}

func cosine(a, b []float32) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		na += float64(a[i]) * float64(a[i])
		nb += float64(b[i]) * float64(b[i])
	}
	denom := math.Sqrt(na) * math.Sqrt(nb)
	if denom == 0 {
		return 0
	}
	return dot / denom
}
