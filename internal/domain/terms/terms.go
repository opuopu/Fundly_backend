package terms

import "time"

type TermsAndConditions struct {
	Version   string    `json:"version"`
	Content   string    `json:"content"`
	Effective time.Time `json:"effective"`
}
