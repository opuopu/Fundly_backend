package budget

import (
	"context"
	"math"

	domainBudget "github.com/Dodul01/fundly/internal/domain/budget"
	domainCategory "github.com/Dodul01/fundly/internal/domain/category"
)

// SpentProvider yields the total EXPENSE amount recorded against a category in a month.
// The postgres transaction repository satisfies this interface.
type SpentProvider interface {
	SpentByCategoryMonth(ctx context.Context, userID, categoryID, month string) (int64, error)
}

type CategoryRef struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
	Icon string `json:"icon,omitempty"`
}

type BudgetResponse struct {
	ID                  string       `json:"id"`
	Category            *CategoryRef `json:"category,omitempty"`
	Month               string       `json:"month"`
	Date                string       `json:"date"`
	LimitAmount         int64        `json:"limit_amount"`
	Expenditure         int64        `json:"expenditure"`
	RemainingAmount     int64        `json:"remaining_amount"`
	UsagePercentage     float64      `json:"usage_percentage"`
	RemainingPercentage float64      `json:"remaining_percentage"`
	Currency            string       `json:"currency"`
	Note                string       `json:"note,omitempty"`
	Description         string       `json:"description,omitempty"`
	CreatedAt           string       `json:"created_at"`
}

// builder resolves the derived spend figures and nested category details for a budget.
type builder struct {
	categories domainCategory.Repository
	spent      SpentProvider
}

func (b builder) toResponse(ctx context.Context, bg *domainBudget.Budget) BudgetResponse {
	var spent int64
	if b.spent != nil {
		if s, err := b.spent.SpentByCategoryMonth(ctx, bg.UserID, bg.CategoryID, bg.Month); err == nil {
			spent = s
		}
	}

	usage := 0.0
	if bg.LimitAmount > 0 {
		usage = math.Round(float64(spent)/float64(bg.LimitAmount)*10000) / 100
	}
	// Remaining is the inverse of usage, clamped to 0 once spending exceeds the limit —
	// a negative "remaining" reads as a bug on the client, not an over-budget signal.
	remaining := math.Max(0, 100-usage)

	res := BudgetResponse{
		ID:                  bg.ID,
		Month:               bg.Month,
		Date:                bg.Date.Format("2006-01-02"),
		LimitAmount:         bg.LimitAmount,
		Expenditure:         spent,
		RemainingAmount:     bg.LimitAmount - spent,
		UsagePercentage:     usage,
		RemainingPercentage: remaining,
		Currency:            bg.Currency,
		Note:                bg.Note,
		Description:         bg.Description,
		CreatedAt:           bg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	ref := &CategoryRef{ID: bg.CategoryID}
	if b.categories != nil {
		if cat, err := b.categories.GetByID(ctx, bg.CategoryID); err == nil && cat != nil {
			ref.Name = cat.Name
			ref.Icon = cat.Icon
		}
	}
	res.Category = ref

	return res
}
