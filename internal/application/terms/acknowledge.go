package terms

import (
	"context"
	"strings"
	"time"

	domainTerms "github.com/Dodul01/fundly/internal/domain/terms"
)

type AcknowledgeInput struct {
	UserID  string `json:"-"`
	Version string `json:"version"`
}

type AcknowledgeUseCase struct {
	repo domainTerms.Repository
}

func NewAcknowledgeUseCase(repo domainTerms.Repository) *AcknowledgeUseCase {
	return &AcknowledgeUseCase{repo: repo}
}

func (uc *AcknowledgeUseCase) Execute(ctx context.Context, input AcknowledgeInput) error {
	version := strings.TrimSpace(input.Version)
	if version == "" {
		return domainTerms.ErrVersionMissing
	}

	// Ensure the referenced version actually exists before recording acceptance.
	if _, err := uc.repo.GetByVersion(ctx, version); err != nil {
		return err
	}

	return uc.repo.SaveAcknowledgment(ctx, &domainTerms.Acknowledgment{
		UserID:         input.UserID,
		Version:        version,
		AcknowledgedAt: time.Now().UTC(),
	})
}
