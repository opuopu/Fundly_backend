package privacypolicy

import (
	"context"
	"fmt"
	"strings"

	domainPrivacyPolicy "github.com/Dodul01/fundly/internal/domain/privacypolicy"
)

type Clause struct {
	ClauseNumber string `json:"clause_number"`
	Title        string `json:"title"`
	Content      string `json:"content"`
}

type LatestPrivacyPolicyResponse struct {
	Version       string   `json:"version"`
	EffectiveDate string   `json:"effective_date"`
	Title         string   `json:"title"`
	Clauses       []Clause `json:"clauses"`
}

type GetLatestUseCase struct {
	repo domainPrivacyPolicy.Repository
}

func NewGetLatestUseCase(repo domainPrivacyPolicy.Repository) *GetLatestUseCase {
	return &GetLatestUseCase{repo: repo}
}

func (uc *GetLatestUseCase) Execute(ctx context.Context) (*LatestPrivacyPolicyResponse, error) {
	p, err := uc.repo.GetLatest(ctx)
	if err != nil {
		return nil, err
	}

	return &LatestPrivacyPolicyResponse{
		Version:       p.Version,
		EffectiveDate: p.Effective.Format("2006-01-02"),
		Title:         "Privacy Policy",
		Clauses:       splitClauses(p.Content),
	}, nil
}

// splitClauses turns the stored content into numbered clauses. Paragraphs are separated by
// blank lines; a single block yields one clause.
func splitClauses(content string) []Clause {
	content = strings.TrimSpace(content)
	if content == "" {
		return []Clause{}
	}

	parts := strings.Split(content, "\n\n")
	clauses := make([]Clause, 0, len(parts))
	n := 0
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n++
		clauses = append(clauses, Clause{
			ClauseNumber: fmt.Sprintf("%02d", n),
			Title:        fmt.Sprintf("Clause %02d", n),
			Content:      p,
		})
	}
	return clauses
}
