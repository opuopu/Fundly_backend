package feedback_test

import (
	"context"
	"errors"
	"testing"

	feedbackapp "github.com/Dodul01/fundly/internal/application/feedback"
	domainFeedback "github.com/Dodul01/fundly/internal/domain/feedback"
)

type mockFeedbackRepo struct {
	items     []*domainFeedback.Feedback
	createErr error
}

func (m *mockFeedbackRepo) Create(_ context.Context, f *domainFeedback.Feedback) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.items = append(m.items, f)
	return nil
}

func TestSubmitFeedback_Success(t *testing.T) {
	repo := &mockFeedbackRepo{}
	uc := feedbackapp.NewSubmitFeedbackUseCase(repo)

	res, err := uc.Execute(context.Background(), feedbackapp.SubmitFeedbackInput{
		UserID:      "usr_1",
		Subject:     "App Performance Feedback",
		Description: "Transactions sync immediately!",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID == "" {
		t.Error("expected generated feedback ID")
	}
	if len(repo.items) != 1 || repo.items[0].UserID != "usr_1" {
		t.Errorf("expected feedback persisted for user, got %+v", repo.items)
	}
}

func TestSubmitFeedback_EmptySubject(t *testing.T) {
	uc := feedbackapp.NewSubmitFeedbackUseCase(&mockFeedbackRepo{})
	_, err := uc.Execute(context.Background(), feedbackapp.SubmitFeedbackInput{
		UserID:      "usr_1",
		Subject:     "   ",
		Description: "something",
	})
	if !errors.Is(err, domainFeedback.ErrInvalidSubject) {
		t.Fatalf("expected ErrInvalidSubject, got %v", err)
	}
}

func TestSubmitFeedback_EmptyDescription(t *testing.T) {
	uc := feedbackapp.NewSubmitFeedbackUseCase(&mockFeedbackRepo{})
	_, err := uc.Execute(context.Background(), feedbackapp.SubmitFeedbackInput{
		UserID:  "usr_1",
		Subject: "Subject",
	})
	if !errors.Is(err, domainFeedback.ErrInvalidDescription) {
		t.Fatalf("expected ErrInvalidDescription, got %v", err)
	}
}
