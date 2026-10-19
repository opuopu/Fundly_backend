package report

import (
	"context"
	"errors"
	"strings"
	"time"

	domainBudget "github.com/Dodul01/fundly/internal/domain/budget"
	domainCategory "github.com/Dodul01/fundly/internal/domain/category"
	domainConfig "github.com/Dodul01/fundly/internal/domain/config"
	domainStat "github.com/Dodul01/fundly/internal/domain/stat"
)

// ConfigReader resolves a user's preferred currency. The config repository satisfies it.
type ConfigReader interface {
	Get(ctx context.Context, userID string) (*domainConfig.Config, error)
}

var (
	ErrInvalidType  = errors.New("type must be INCOME or EXPENSE")
	ErrInvalidMonth = errors.New("month must be in YYYY-MM format")
)

type CompareInput struct {
	UserID          string
	Type            string
	PrimaryMonth    string
	ComparisonMonth string
	CategoryIDs     []string
}

type CategoryComparison struct {
	CategoryID       string `json:"category_id"`
	Name             string `json:"name,omitempty"`
	PrimaryAmount    int64  `json:"primary_amount"`
	ComparisonAmount int64  `json:"comparison_amount"`
	Difference       int64  `json:"difference"`
}

type CompareResponse struct {
	Type             string               `json:"type"`
	PrimaryPeriod    string               `json:"primary_period"`
	ComparisonPeriod string               `json:"comparison_period"`
	ComparedBalance  int64                `json:"compared_balance"`
	Currency         string               `json:"currency"`
	PrimaryTotal     int64                `json:"primary_total"`
	ComparisonTotal  int64                `json:"comparison_total"`
	Categories       []CategoryComparison `json:"categories"`
}

type CompareUseCase struct {
	stats   domainStat.Repository
	cats    domainCategory.Repository
	configs ConfigReader
}

func NewCompareUseCase(stats domainStat.Repository, cats domainCategory.Repository, configs ConfigReader) *CompareUseCase {
	return &CompareUseCase{stats: stats, cats: cats, configs: configs}
}

func (uc *CompareUseCase) Execute(ctx context.Context, input CompareInput) (*CompareResponse, error) {
	txType, err := normalizeType(input.Type)
	if err != nil {
		return nil, err
	}

	primaryMonth := strings.TrimSpace(input.PrimaryMonth)
	if primaryMonth == "" {
		primaryMonth = time.Now().UTC().Format("2006-01")
	}
	if !domainBudget.ValidMonth(primaryMonth) {
		return nil, ErrInvalidMonth
	}

	comparisonMonth := strings.TrimSpace(input.ComparisonMonth)
	if comparisonMonth == "" {
		comparisonMonth = previousMonth(primaryMonth)
	}
	if !domainBudget.ValidMonth(comparisonMonth) {
		return nil, ErrInvalidMonth
	}

	primary, err := uc.stats.BreakdownByCategory(ctx, input.UserID, txType, primaryMonth, "")
	if err != nil {
		return nil, err
	}
	comparison, err := uc.stats.BreakdownByCategory(ctx, input.UserID, txType, comparisonMonth, "")
	if err != nil {
		return nil, err
	}

	primaryByCat := toMap(primary)
	comparisonByCat := toMap(comparison)

	// Determine which categories to report: an explicit list (preserving its order) or the
	// union of categories seen in either period.
	categoryIDs := cleanIDs(input.CategoryIDs)
	if len(categoryIDs) == 0 {
		categoryIDs = unionKeys(primary, comparison)
	}

	currency := domainConfig.DefaultCurrency
	if uc.configs != nil {
		if c, err := uc.configs.Get(ctx, input.UserID); err == nil && c != nil && c.DefaultCurrency != "" {
			currency = c.DefaultCurrency
		}
	}

	res := &CompareResponse{
		Type:             txType,
		PrimaryPeriod:    primaryMonth,
		ComparisonPeriod: comparisonMonth,
		Currency:         currency,
		Categories:       make([]CategoryComparison, 0, len(categoryIDs)),
	}

	for _, id := range categoryIDs {
		p := primaryByCat[id]
		c := comparisonByCat[id]
		cmp := CategoryComparison{
			CategoryID:       id,
			PrimaryAmount:    p,
			ComparisonAmount: c,
			Difference:       p - c,
		}
		if uc.cats != nil {
			if cat, err := uc.cats.GetByID(ctx, id); err == nil && cat != nil {
				cmp.Name = cat.Name
			}
		}
		res.Categories = append(res.Categories, cmp)
		res.PrimaryTotal += p
		res.ComparisonTotal += c
	}

	res.ComparedBalance = res.PrimaryTotal - res.ComparisonTotal
	return res, nil
}

func normalizeType(s string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "INCOME":
		return "INCOME", nil
	case "EXPENSE":
		return "EXPENSE", nil
	default:
		return "", ErrInvalidType
	}
}

func toMap(slices []domainStat.CategoryAmount) map[string]int64 {
	m := make(map[string]int64, len(slices))
	for _, s := range slices {
		m[s.CategoryID] = s.Amount
	}
	return m
}

// unionKeys returns category IDs present in either period, primary order first.
func unionKeys(primary, comparison []domainStat.CategoryAmount) []string {
	seen := make(map[string]bool)
	var out []string
	for _, s := range primary {
		if !seen[s.CategoryID] {
			seen[s.CategoryID] = true
			out = append(out, s.CategoryID)
		}
	}
	for _, s := range comparison {
		if !seen[s.CategoryID] {
			seen[s.CategoryID] = true
			out = append(out, s.CategoryID)
		}
	}
	return out
}

func cleanIDs(ids []string) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// previousMonth returns the YYYY-MM immediately before the given YYYY-MM.
func previousMonth(month string) string {
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return month
	}
	return t.AddDate(0, -1, 0).Format("2006-01")
}
