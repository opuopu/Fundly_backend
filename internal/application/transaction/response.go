package transaction

import (
	"context"
	"strings"
	"time"

	domainAccount "github.com/Dodul01/fundly/internal/domain/account"
	domainAttachment "github.com/Dodul01/fundly/internal/domain/attachment"
	domainCategory "github.com/Dodul01/fundly/internal/domain/category"
	domainTx "github.com/Dodul01/fundly/internal/domain/transaction"
)

// AttachmentProvider supplies attachment data for a transaction. The postgres attachment
// repository satisfies it. It is optional on every use case (may be nil).
type AttachmentProvider interface {
	ListByTransactionID(ctx context.Context, transactionID string) ([]*domainAttachment.Attachment, error)
	CountByTransactionID(ctx context.Context, transactionID string) (int, error)
	LinkToTransaction(ctx context.Context, userID, transactionID string, attachmentIDs []string) error
}

// CategoryRef is the nested category summary embedded in a transaction response.
type CategoryRef struct {
	ID    string `json:"id"`
	Name  string `json:"name,omitempty"`
	Icon  string `json:"icon,omitempty"`
	Color string `json:"color,omitempty"`
}

// AccountRef is the nested account summary embedded in a transaction response.
type AccountRef struct {
	ID    string `json:"id"`
	Name  string `json:"name,omitempty"`
	Group string `json:"group,omitempty"`
}

// AttachmentRef is the nested attachment summary embedded in a transaction detail response.
type AttachmentRef struct {
	ID       string `json:"id"`
	FileName string `json:"file_name"`
	URL      string `json:"url"`
	FileSize int64  `json:"file_size"`
}

type TransactionResponse struct {
	ID              string          `json:"id"`
	Type            string          `json:"type"`
	Date            time.Time       `json:"date"`
	Amount          int64           `json:"amount"`
	Currency        string          `json:"currency"`
	Category        *CategoryRef    `json:"category,omitempty"`
	Account         *AccountRef     `json:"account,omitempty"`
	Note            string          `json:"note,omitempty"`
	Description     string          `json:"description,omitempty"`
	AttachmentCount int             `json:"attachment_count"`
	Attachments     []AttachmentRef `json:"attachments,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
}

// PaginationMeta mirrors the HTTP pagination envelope so handlers can map it directly.
type PaginationMeta struct {
	Page         int
	Limit        int
	TotalRecords int
	TotalPages   int
}

// enricher resolves nested category/account/attachment details. Any provider may be nil, in
// which case only the referenced IDs (or a zero count) are returned.
type enricher struct {
	categories  domainCategory.Repository
	accounts    domainAccount.Repository
	attachments AttachmentProvider
}

// toResponse builds a list-level response (attachment count only).
func (e enricher) toResponse(ctx context.Context, tx *domainTx.Transaction) TransactionResponse {
	res := e.base(ctx, tx)
	if e.attachments != nil {
		if n, err := e.attachments.CountByTransactionID(ctx, tx.ID); err == nil {
			res.AttachmentCount = n
		}
	}
	return res
}

// toDetailResponse builds a detail-level response including the full attachment list.
func (e enricher) toDetailResponse(ctx context.Context, tx *domainTx.Transaction) TransactionResponse {
	res := e.base(ctx, tx)
	if e.attachments != nil {
		if atts, err := e.attachments.ListByTransactionID(ctx, tx.ID); err == nil {
			res.AttachmentCount = len(atts)
			res.Attachments = make([]AttachmentRef, 0, len(atts))
			for _, a := range atts {
				res.Attachments = append(res.Attachments, AttachmentRef{
					ID:       a.ID,
					FileName: a.FileName,
					URL:      a.FileURL,
					FileSize: a.FileSize,
				})
			}
		}
	}
	return res
}

func (e enricher) base(ctx context.Context, tx *domainTx.Transaction) TransactionResponse {
	res := TransactionResponse{
		ID:          tx.ID,
		Type:        string(tx.Type),
		Date:        tx.Date,
		Amount:      tx.Amount,
		Currency:    tx.Currency,
		Note:        tx.Note,
		Description: tx.Description,
		CreatedAt:   tx.CreatedAt,
	}

	if tx.CategoryID != "" {
		ref := &CategoryRef{ID: tx.CategoryID}
		if e.categories != nil {
			if cat, err := e.categories.GetByID(ctx, tx.CategoryID); err == nil && cat != nil {
				ref.Name = cat.Name
				ref.Icon = cat.Icon
				ref.Color = cat.Color
			}
		}
		res.Category = ref
	}

	if accID := strings.TrimSpace(tx.AccountID()); accID != "" {
		ref := &AccountRef{ID: accID}
		if e.accounts != nil {
			if acc, err := e.accounts.GetByID(ctx, accID); err == nil && acc != nil {
				ref.Name = acc.Name
				ref.Group = acc.Group
			}
		}
		res.Account = ref
	}

	return res
}
