package transaction

import (
	"context"

	domainAccount "github.com/Dodul01/fundly/internal/domain/account"
	domainCategory "github.com/Dodul01/fundly/internal/domain/category"
	domainTx "github.com/Dodul01/fundly/internal/domain/transaction"
)

type GetTransactionUseCase struct {
	txRepo      domainTx.Repository
	accounts    domainAccount.Repository
	cats        domainCategory.Repository
	attachments AttachmentProvider
}

func NewGetTransactionUseCase(txRepo domainTx.Repository, accounts domainAccount.Repository, cats domainCategory.Repository, attachments AttachmentProvider) *GetTransactionUseCase {
	return &GetTransactionUseCase{txRepo: txRepo, accounts: accounts, cats: cats, attachments: attachments}
}

func (uc *GetTransactionUseCase) Execute(ctx context.Context, userID, id string) (*TransactionResponse, error) {
	tx, err := uc.txRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Ownership guard: a mismatched owner is indistinguishable from not-found.
	if tx.UserID != userID {
		return nil, domainTx.ErrTransactionNotFound
	}

	res := enricher{categories: uc.cats, accounts: uc.accounts, attachments: uc.attachments}.toDetailResponse(ctx, tx)
	return &res, nil
}
