package stat

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	domainCategory "github.com/Dodul01/fundly/internal/domain/category"
	domainConfig "github.com/Dodul01/fundly/internal/domain/config"
	domainStat "github.com/Dodul01/fundly/internal/domain/stat"
)

var ErrInvalidType = errors.New("type must be INCOME or EXPENSE")

// ConfigReader resolves a user's preferred currency. The config repository satisfies it.
type ConfigReader interface {
	Get(ctx context.Context, userID string) (*domainConfig.Config, error)
}

// resolveCurrency returns the user's configured currency, or USD when unavailable.
func resolveCurrency(ctx context.Context, configs ConfigReader, userID string) string {
	if configs != nil {
		if c, err := configs.Get(ctx, userID); err == nil && c != nil && c.DefaultCurrency != "" {
			return c.DefaultCurrency
		}
	}
	return domainConfig.DefaultCurrency
}

// CurrentMonth returns the current UTC month as YYYY-MM.
func CurrentMonth() string {
	return time.Now().UTC().Format("2006-01")
}

// normalizeType validates and upper-cases an INCOME/EXPENSE value.
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

type BreakdownItem struct {
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	Amount       int64   `json:"amount"`
	Percentage   float64 `json:"percentage"`
	Color        string  `json:"color"`
}

type BreakdownResponse struct {
	Type        string          `json:"type"`
	Month       string          `json:"month"`
	TotalAmount int64           `json:"total_amount"`
	Currency    string          `json:"currency"`
	Breakdown   []BreakdownItem `json:"breakdown"`
}

type BreakdownInput struct {
	UserID    string
	Type      string
	Month     string
	AccountID string
}

type BreakdownUseCase struct {
	stats   domainStat.Repository
	cats    domainCategory.Repository
	configs ConfigReader
}

func NewBreakdownUseCase(stats domainStat.Repository, cats domainCategory.Repository, configs ConfigReader) *BreakdownUseCase {
	return &BreakdownUseCase{stats: stats, cats: cats, configs: configs}
}

func (uc *BreakdownUseCase) Execute(ctx context.Context, input BreakdownInput) (*BreakdownResponse, error) {
	txType, err := normalizeType(input.Type)
	if err != nil {
		return nil, err
	}

	month := strings.TrimSpace(input.Month)
	if month == "" {
		month = CurrentMonth()
	}

	slices, err := uc.stats.BreakdownByCategory(ctx, input.UserID, txType, month, input.AccountID)
	if err != nil {
		return nil, err
	}

	var total int64
	for _, s := range slices {
		total += s.Amount
	}

	items := make([]BreakdownItem, 0, len(slices))
	for _, s := range slices {
		item := BreakdownItem{
			CategoryID: s.CategoryID,
			Amount:     s.Amount,
		}
		if total > 0 {
			item.Percentage = math.Round(float64(s.Amount)/float64(total)*10000) / 100
		}
		if uc.cats != nil {
			if cat, err := uc.cats.GetByID(ctx, s.CategoryID); err == nil && cat != nil {
				item.CategoryName = cat.Name
				item.Color = cat.Color
			}
		}
		items = append(items, item)
	}

	return &BreakdownResponse{
		Type:        txType,
		Month:       month,
		TotalAmount: total,
		Currency:    resolveCurrency(ctx, uc.configs, input.UserID),
		Breakdown:   items,
	}, nil
}
