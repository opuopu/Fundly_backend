package privacypolicy_test

import (
	"context"
	"errors"
	"testing"
	"time"

	privacypolicyapp "github.com/Dodul01/fundly/internal/application/privacypolicy"
	domainPrivacyPolicy "github.com/Dodul01/fundly/internal/domain/privacypolicy"
)

type mockPrivacyPolicyRepo struct {
	latest    *domainPrivacyPolicy.PrivacyPolicy
	byVer     map[string]*domainPrivacyPolicy.PrivacyPolicy
	acks      []*domainPrivacyPolicy.Acknowledgment
	latestErr error
}

func (m *mockPrivacyPolicyRepo) GetLatest(_ context.Context) (*domainPrivacyPolicy.PrivacyPolicy, error) {
	if m.latestErr != nil {
		return nil, m.latestErr
	}
	if m.latest == nil {
		return nil, domainPrivacyPolicy.ErrPolicyNotFound
	}
	return m.latest, nil
}

func (m *mockPrivacyPolicyRepo) GetByVersion(_ context.Context, version string) (*domainPrivacyPolicy.PrivacyPolicy, error) {
	p, ok := m.byVer[version]
	if !ok {
		return nil, domainPrivacyPolicy.ErrPolicyNotFound
	}
	return p, nil
}

func (m *mockPrivacyPolicyRepo) SaveAcknowledgment(_ context.Context, a *domainPrivacyPolicy.Acknowledgment) error {
	m.acks = append(m.acks, a)
	return nil
}

func TestGetLatest_BuildsClauses(t *testing.T) {
	repo := &mockPrivacyPolicyRepo{latest: &domainPrivacyPolicy.PrivacyPolicy{
		Version:   "1.0",
		Content:   "First clause content.\n\nSecond clause content.",
		Effective: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}}
	uc := privacypolicyapp.NewGetLatestUseCase(repo)

	res, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Version != "1.0" || res.EffectiveDate != "2026-01-01" {
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
	uc := privacypolicyapp.NewGetLatestUseCase(&mockPrivacyPolicyRepo{})
	_, err := uc.Execute(context.Background())
	if !errors.Is(err, domainPrivacyPolicy.ErrPolicyNotFound) {
		t.Fatalf("expected ErrPolicyNotFound, got %v", err)
	}
}

func TestAcknowledge_Success(t *testing.T) {
	repo := &mockPrivacyPolicyRepo{byVer: map[string]*domainPrivacyPolicy.PrivacyPolicy{
		"1.0": {Version: "1.0"},
	}}
	uc := privacypolicyapp.NewAcknowledgeUseCase(repo)

	if err := uc.Execute(context.Background(), privacypolicyapp.AcknowledgeInput{UserID: "usr_1", Version: "1.0"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.acks) != 1 || repo.acks[0].UserID != "usr_1" || repo.acks[0].Version != "1.0" {
		t.Errorf("expected acknowledgment recorded, got %+v", repo.acks)
	}
}

func TestAcknowledge_MissingVersion(t *testing.T) {
	uc := privacypolicyapp.NewAcknowledgeUseCase(&mockPrivacyPolicyRepo{})
	err := uc.Execute(context.Background(), privacypolicyapp.AcknowledgeInput{UserID: "usr_1", Version: "  "})
	if !errors.Is(err, domainPrivacyPolicy.ErrVersionMissing) {
		t.Fatalf("expected ErrVersionMissing, got %v", err)
	}
}

func TestAcknowledge_UnknownVersion(t *testing.T) {
	uc := privacypolicyapp.NewAcknowledgeUseCase(&mockPrivacyPolicyRepo{byVer: map[string]*domainPrivacyPolicy.PrivacyPolicy{}})
	err := uc.Execute(context.Background(), privacypolicyapp.AcknowledgeInput{UserID: "usr_1", Version: "9.9"})
	if !errors.Is(err, domainPrivacyPolicy.ErrPolicyNotFound) {
		t.Fatalf("expected ErrPolicyNotFound, got %v", err)
	}
}
