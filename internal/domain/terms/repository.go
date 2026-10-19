package terms

import "context"

type Repository interface {
	// GetLatest returns the most recent terms by effective date.
	GetLatest(ctx context.Context) (*TermsAndConditions, error)
	// GetByVersion returns the terms for a specific version.
	GetByVersion(ctx context.Context, version string) (*TermsAndConditions, error)
	// SaveAcknowledgment records (idempotently) a user's acceptance of a terms version.
	SaveAcknowledgment(ctx context.Context, a *Acknowledgment) error
}
