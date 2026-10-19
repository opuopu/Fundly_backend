package transaction

import (
	"context"
	"strings"

	domainAccount "github.com/Dodul01/fundly/internal/domain/account"
	domainCategory "github.com/Dodul01/fundly/internal/domain/category"
	domainTx "github.com/Dodul01/fundly/internal/domain/transaction"
)

type SearchTransactionsUseCase struct {
	txRepo      domainTx.Repository
	accounts    domainAccount.Repository
	cats        domainCategory.Repository
	attachments AttachmentProvider
}

func NewSearchTransactionsUseCase(txRepo domainTx.Repository, accounts domainAccount.Repository, cats domainCategory.Repository, attachments AttachmentProvider) *SearchTransactionsUseCase {
	return &SearchTransactionsUseCase{txRepo: txRepo, accounts: accounts, cats: cats, attachments: attachments}
}

func (uc *SearchTransactionsUseCase) Execute(ctx context.Context, userID, query string, limit int) ([]TransactionResponse, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []TransactionResponse{}, nil
	}

	txs, err := uc.txRepo.Search(ctx, userID, query, limit)
	if err != nil {
		return nil, err
	}

	en := enricher{categories: uc.cats, accounts: uc.accounts, attachments: uc.attachments}
	res := make([]TransactionResponse, 0, len(txs))
	for _, tx := range txs {
		res = append(res, en.toResponse(ctx, tx))
	}

	return res, nil
}
