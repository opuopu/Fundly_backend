package privacypolicy

import (
	"context"
	"strings"
	"time"

	domainPrivacyPolicy "github.com/Dodul01/fundly/internal/domain/privacypolicy"
)

type AcknowledgeInput struct {
	UserID  string `json:"-"`
	Version string `json:"version"`
}

type AcknowledgeUseCase struct {
	repo domainPrivacyPolicy.Repository
}

func NewAcknowledgeUseCase(repo domainPrivacyPolicy.Repository) *AcknowledgeUseCase {
	return &AcknowledgeUseCase{repo: repo}
}

func (uc *AcknowledgeUseCase) Execute(ctx context.Context, input AcknowledgeInput) error {
	version := strings.TrimSpace(input.Version)
	if version == "" {
		return domainPrivacyPolicy.ErrVersionMissing
	}

	// Ensure the referenced version actually exists before recording acceptance.
	if _, err := uc.repo.GetByVersion(ctx, version); err != nil {
		return err
	}

	return uc.repo.SaveAcknowledgment(ctx, &domainPrivacyPolicy.Acknowledgment{
		UserID:         input.UserID,
		Version:        version,
		AcknowledgedAt: time.Now().UTC(),
	})
}
