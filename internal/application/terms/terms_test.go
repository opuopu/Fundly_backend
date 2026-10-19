package terms_test

import (
	"context"
	"errors"
	"testing"
	"time"

	termsapp "github.com/Dodul01/fundly/internal/application/terms"
	domainTerms "github.com/Dodul01/fundly/internal/domain/terms"
)

type mockTermsRepo struct {
	latest    *domainTerms.TermsAndConditions
	byVer     map[string]*domainTerms.TermsAndConditions
	acks      []*domainTerms.Acknowledgment
	latestErr error
}

func (m *mockTermsRepo) GetLatest(_ context.Context) (*domainTerms.TermsAndConditions, error) {
	if m.latestErr != nil {
		return nil, m.latestErr
	}
	if m.latest == nil {
		return nil, domainTerms.ErrTermsNotFound
	}
	return m.latest, nil
}

func (m *mockTermsRepo) GetByVersion(_ context.Context, version string) (*domainTerms.TermsAndConditions, error) {
	t, ok := m.byVer[version]
	if !ok {
		return nil, domainTerms.ErrTermsNotFound
	}
	return t, nil
}

func (m *mockTermsRepo) SaveAcknowledgment(_ context.Context, a *domainTerms.Acknowledgment) error {
	m.acks = append(m.acks, a)
	return nil
}

func TestGetLatest_BuildsClauses(t *testing.T) {
	repo := &mockTermsRepo{latest: &domainTerms.TermsAndConditions{
		Version:   "2.1",
		Content:   "First clause content.\n\nSecond clause content.",
		Effective: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}}
	uc := termsapp.NewGetLatestUseCase(repo)

	res, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Version != "2.1" || res.EffectiveDate != "2026-01-01" {
		t.Errorf("unexpected version/date: %+v", res)
	}
	if len(res.Clauses) != 2 {
		t.Fatalf("expected 2 clauses, got %d", len(res.Clauses))
	}
	if res.Clauses[0].ClauseNumber != "01" || res.Clauses[1].ClauseNumber != "02" {
		t.Errorf("unexpected clause numbering: %+v", res.Clauses)
	}
}

func TestGetLatest_NotFound(t *testing.T) {
	uc := termsapp.NewGetLatestUseCase(&mockTermsRepo{})
	_, err := uc.Execute(context.Background())
	if !errors.Is(err, domainTerms.ErrTermsNotFound) {
		t.Fatalf("expected ErrTermsNotFound, got %v", err)
	}
}

func TestAcknowledge_Success(t *testing.T) {
	repo := &mockTermsRepo{byVer: map[string]*domainTerms.TermsAndConditions{
		"2.1": {Version: "2.1"},
	}}
	uc := termsapp.NewAcknowledgeUseCase(repo)

	if err := uc.Execute(context.Background(), termsapp.AcknowledgeInput{UserID: "usr_1", Version: "2.1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.acks) != 1 || repo.acks[0].UserID != "usr_1" || repo.acks[0].Version != "2.1" {
		t.Errorf("expected acknowledgment recorded, got %+v", repo.acks)
	}
}

func TestAcknowledge_MissingVersion(t *testing.T) {
	uc := termsapp.NewAcknowledgeUseCase(&mockTermsRepo{})
	err := uc.Execute(context.Background(), termsapp.AcknowledgeInput{UserID: "usr_1", Version: "  "})
	if !errors.Is(err, domainTerms.ErrVersionMissing) {
		t.Fatalf("expected ErrVersionMissing, got %v", err)
	}
}

func TestAcknowledge_UnknownVersion(t *testing.T) {
	uc := termsapp.NewAcknowledgeUseCase(&mockTermsRepo{byVer: map[string]*domainTerms.TermsAndConditions{}})
	err := uc.Execute(context.Background(), termsapp.AcknowledgeInput{UserID: "usr_1", Version: "9.9"})
	if !errors.Is(err, domainTerms.ErrTermsNotFound) {
		t.Fatalf("expected ErrTermsNotFound, got %v", err)
	}
}
