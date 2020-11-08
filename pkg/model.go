package pkg

import "time"

// Error .
type Error struct {
	Resource string `json:"resource"`
	Field    string `json:"field"`
	Code     string `json:"code"`
}

// Fault .
type Fault struct {
	Message string   `json:"message"`
	Errors  []*Error `json:"errors"`
}

func (f *Fault) Error() string {
	return f.Message
}

// Export .
type Export struct {
	ID             int        `json:"id"`
	SiteID         int        `json:"site_id"`
	StartFromHitID int        `json:"start_from_hit_id"`
	LastHitID      int        `json:"last_hit_id"`
	Path           string     `json:"path"`
	CreatedAt      time.Time  `json:"created_at"`
	FinishedAt     *time.Time `json:"finished_at"`
	NumRows        int        `json:"num_rows"`
	Size           string     `json:"size"`
	Hash           string     `json:"hash"`
	Error          string     `json:"error"`
}

// Stats .
type Stats struct{}
