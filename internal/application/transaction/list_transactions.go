package transaction

import (
	"context"

	domainAccount "github.com/Dodul01/fundly/internal/domain/account"
	domainCategory "github.com/Dodul01/fundly/internal/domain/category"
	domainTx "github.com/Dodul01/fundly/internal/domain/transaction"
)

type ListTransactionsInput struct {
	UserID     string
	Type       string
	CategoryID string
	AccountID  string
	Date       string
	StartDate  string
	EndDate    string
	Page       int
	Limit      int
}

type ListTransactionsUseCase struct {
	txRepo      domainTx.Repository
	accounts    domainAccount.Repository
	cats        domainCategory.Repository
	attachments AttachmentProvider
}

func NewListTransactionsUseCase(txRepo domainTx.Repository, accounts domainAccount.Repository, cats domainCategory.Repository, attachments AttachmentProvider) *ListTransactionsUseCase {
	return &ListTransactionsUseCase{txRepo: txRepo, accounts: accounts, cats: cats, attachments: attachments}
}

func (uc *ListTransactionsUseCase) Execute(ctx context.Context, input ListTransactionsInput) ([]TransactionResponse, PaginationMeta, error) {
	page := input.Page
	if page < 1 {
		page = 1
	}
	limit := input.Limit
	if limit < 1 {
		limit = 20
	}

	txs, total, err := uc.txRepo.List(ctx, domainTx.Filter{
		UserID:     input.UserID,
		Type:       input.Type,
		CategoryID: input.CategoryID,
		AccountID:  input.AccountID,
		Date:       input.Date,
		StartDate:  input.StartDate,
		EndDate:    input.EndDate,
		Page:       page,
		Limit:      limit,
	})
	if err != nil {
		return nil, PaginationMeta{}, err
	}

	en := enricher{categories: uc.cats, accounts: uc.accounts, attachments: uc.attachments}
	res := make([]TransactionResponse, 0, len(txs))
	for _, tx := range txs {
		res = append(res, en.toResponse(ctx, tx))
	}

	totalPages := 0
	if limit > 0 {
		totalPages = (total + limit - 1) / limit
	}

	meta := PaginationMeta{
		Page:         page,
		Limit:        limit,
		TotalRecords: total,
		TotalPages:   totalPages,
	}

	return res, meta, nil
}
