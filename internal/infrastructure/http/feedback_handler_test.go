package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	feedbackapp "github.com/Dodul01/fundly/internal/application/feedback"
	domainFeedback "github.com/Dodul01/fundly/internal/domain/feedback"
	infrahttp "github.com/Dodul01/fundly/internal/infrastructure/http"
)

type feedbackRepoStub struct {
	items []*domainFeedback.Feedback
}

func (m *feedbackRepoStub) Create(_ context.Context, f *domainFeedback.Feedback) error {
	m.items = append(m.items, f)
	return nil
}

func setupFeedbackHandler() (*infrahttp.FeedbackHandler, *feedbackRepoStub) {
	repo := &feedbackRepoStub{}
	return infrahttp.NewFeedbackHandler(feedbackapp.NewSubmitFeedbackUseCase(repo)), repo
}

func TestFeedbackHandler_Submit(t *testing.T) {
	handler, repo := setupFeedbackHandler()
	body := `{"subject":"Great app","description":"Loving the budgets feature"}`
	req := authed(httptest.NewRequest(http.MethodPost, "/api/v1/feedback", strings.NewReader(body)), "usr_1")
	rec := httptest.NewRecorder()

	handler.SubmitFeedback(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. body: %s", rec.Code, rec.Body.String())
	}
	if len(repo.items) != 1 {
		t.Errorf("expected feedback persisted, got %d", len(repo.items))
	}
}

func TestFeedbackHandler_Submit_EmptySubject(t *testing.T) {
	handler, _ := setupFeedbackHandler()
	req := authed(httptest.NewRequest(http.MethodPost, "/api/v1/feedback", strings.NewReader(`{"description":"x"}`)), "usr_1")
	rec := httptest.NewRecorder()

	handler.SubmitFeedback(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestFeedbackHandler_Unauthorized(t *testing.T) {
	handler, _ := setupFeedbackHandler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/feedback", strings.NewReader(`{"subject":"x","description":"y"}`))
	rec := httptest.NewRecorder()

	handler.SubmitFeedback(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
