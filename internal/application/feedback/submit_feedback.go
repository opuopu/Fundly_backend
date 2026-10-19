package feedback

import (
	"context"
	"strings"

	domainFeedback "github.com/Dodul01/fundly/internal/domain/feedback"
	"github.com/google/uuid"
)

type SubmitFeedbackInput struct {
	UserID      string `json:"-"`
	Subject     string `json:"subject"`
	Description string `json:"description"`
}

type FeedbackResponse struct {
	ID        string `json:"id"`
	Subject   string `json:"subject"`
	CreatedAt string `json:"created_at"`
}

type SubmitFeedbackUseCase struct {
	repo domainFeedback.Repository
}

func NewSubmitFeedbackUseCase(repo domainFeedback.Repository) *SubmitFeedbackUseCase {
	return &SubmitFeedbackUseCase{repo: repo}
}

func (uc *SubmitFeedbackUseCase) Execute(ctx context.Context, input SubmitFeedbackInput) (*FeedbackResponse, error) {
	feedbackID := "fb_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:8]

	f, err := domainFeedback.New(feedbackID, input.UserID, input.Subject, input.Description)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Create(ctx, f); err != nil {
		return nil, err
	}

	return &FeedbackResponse{
		ID:        f.ID,
		Subject:   f.Subject,
		CreatedAt: f.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
