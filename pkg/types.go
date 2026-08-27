package calderadb

import "time"

// Document represents a document in CalderaDB
type Document struct {
	ID         string                 `json:"id"`
	Collection string                 `json:"collection"`
	Data       map[string]interface{} `json:"data"`
	Version    int                    `json:"version"`
	CreatedAt  time.Time              `json:"created_at"`
	ModifiedAt time.Time              `json:"modified_at"`
	Location   string                 `json:"location"` // "hot" or "cold"
	Accesses   int                    `json:"accesses"`
}

// Stats represents database statistics
type Stats struct {
	TotalGets int64   `json:"total_gets"`
	TotalSets int64   `json:"total_sets"`
	TotalDels int64   `json:"total_dels"`
	HotHits   int64   `json:"hot_hits"`
	ColdHits  int64   `json:"cold_hits"`
	Misses    int64   `json:"misses"`
	HotDocs   int64   `json:"hot_docs"`
	HotBytes  int64   `json:"hot_bytes"`
	ColdDocs  int64   `json:"cold_docs"`
	ColdBytes int64   `json:"cold_bytes"`
	HitRate   float64 `json:"hit_rate"`
}

// CollectionInfo represents collection information
type CollectionInfo struct {
	Name       string    `json:"name"`
	DocCount   int64     `json:"doc_count"`
	HotDocs    int64     `json:"hot_docs"`
	ColdDocs   int64     `json:"cold_docs"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}

// TierDocument represents a document in a specific tier
type TierDocument struct {
	ID       string                 `json:"id"`
	Data     map[string]interface{} `json:"data"`
	Accesses int                    `json:"accesses,omitempty"`
	Offset   int64                  `json:"offset,omitempty"`
}
