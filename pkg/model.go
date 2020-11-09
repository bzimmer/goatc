package pkg

import "time"

// Fault .
type Fault struct {
	Err  string              `json:"error"`
	Errs map[string][]string `json:"errors"`
}

func (f *Fault) Error() string {
	return f.Err
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
type Stats struct {
	Path           string    `json:"path"`
	Title          string    `json:"title"`
	Event          bool      `json:"event"`
	Bot            int       `json:"bot"`
	Session        string    `json:"session"`
	FirstVisit     bool      `json:"first_visit"`
	Referrer       string    `json:"referrer"`
	ReferrerScheme string    `json:"referrer_scheme"`
	UserAgent      string    `json:"user_agent"`
	ScreenSize     string    `json:"screen_size"`
	Location       string    `json:"location"`
	Date           time.Time `json:"date"`
}

// ExportedStats .
type ExportedStats struct {
	*Export `json:"export"`
	Stats   []*Stats `json:"stats"`
}
