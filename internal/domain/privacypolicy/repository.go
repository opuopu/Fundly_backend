package privacypolicy

import "context"

type Repository interface {
	GetLatest(ctx context.Context) (*PrivacyPolicy, error)
	GetByVersion(ctx context.Context, version string) (*PrivacyPolicy, error)
	SaveAcknowledgment(ctx context.Context, a *Acknowledgment) error
}
