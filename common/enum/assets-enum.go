package enum

import "sync"

type AssetID string

const (
	USD AssetID = "USD"
	BTC AssetID = "BTC"
	ETH AssetID = "ETH"
)

type AssetsMeta struct {
	ID        AssetID
	Precision uint32
	Active    bool
}

type AssetRegistry interface {
	Exists(id AssetID) bool
	Get(id AssetID) (AssetsMeta, bool)
}

type MemoryRegistry struct {
	mu    sync.RWMutex
	metas map[AssetID]AssetsMeta
}

func NewMemoryRegistry() *MemoryRegistry {
	r := &MemoryRegistry{metas: make(map[AssetID]AssetsMeta)}
    r.Register(AssetsMeta{ID: USD, Precision: 2, Active: true})
    r.Register(AssetsMeta{ID: BTC, Precision: 8, Active: true})
    r.Register(AssetsMeta{ID: ETH, Precision: 8, Active: true})
    return r
}

func (r *MemoryRegistry) Register(m AssetsMeta) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metas[m.ID] = m
}

func (r *MemoryRegistry) Exists(id AssetID) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.metas[id]
	return ok && m.Active
}

func (r *MemoryRegistry) Get(id AssetID) (AssetsMeta, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.metas[id]
	return m, ok
}
