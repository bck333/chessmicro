package engine

import (
	"sync"
	"time"
)

// analysisCacheEntry stores a cached AnalysisResult with an expiry time.
type analysisCacheEntry struct {
	result  *AnalysisResult
	expires time.Time
}

// AnalysisCache is a bounded, TTL-based cache for analysis results keyed by FEN.
// It evicts the oldest entries when the max size is reached and expires entries by TTL.
type AnalysisCache struct {
	mu       sync.Mutex
	entries  map[string]*analysisCacheEntry
	order    []string // FEN keys in insertion order for bounded eviction
	maxItems int
	ttl      time.Duration
}

// NewAnalysisCache creates a bounded TTL cache.
func NewAnalysisCache(maxItems int, ttl time.Duration) *AnalysisCache {
	if maxItems < 1 {
		maxItems = 256
	}
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &AnalysisCache{
		entries:  make(map[string]*analysisCacheEntry),
		maxItems: maxItems,
		ttl:      ttl,
	}
}

// Load returns a cached result if present and not expired, plus whether it is usable
// for the requested depth and MultiPV count.
func (c *AnalysisCache) Load(fen string, minDepth int, minLines int) (*AnalysisResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[fen]
	if !ok {
		return nil, false
	}
	if time.Now().After(e.expires) {
		delete(c.entries, fen)
		c.removeFromOrder(fen)
		return nil, false
	}
	if e.result.Depth < minDepth || len(e.result.Lines) < minLines {
		return nil, false
	}
	return e.result, true
}

// Store adds or refreshes a cache entry, evicting the oldest if over capacity.
func (c *AnalysisCache) Store(fen string, result *AnalysisResult) {
	if result == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.entries[fen]; !exists {
		c.order = append(c.order, fen)
		if len(c.order) > c.maxItems {
			oldest := c.order[0]
			c.order = c.order[1:]
			delete(c.entries, oldest)
		}
	}
	c.entries[fen] = &analysisCacheEntry{
		result:  result,
		expires: time.Now().Add(c.ttl),
	}
}

// Size returns the current number of cached entries (for observability/admin).
func (c *AnalysisCache) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.entries)
}

func (c *AnalysisCache) removeFromOrder(fen string) {
	for i, k := range c.order {
		if k == fen {
			c.order = append(c.order[:i], c.order[i+1:]...)
			return
		}
	}
}
