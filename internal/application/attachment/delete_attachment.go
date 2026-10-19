package attachment

import (
	"context"

	domainAttachment "github.com/Dodul01/fundly/internal/domain/attachment"
)

type DeleteAttachmentUseCase struct {
	repo domainAttachment.Repository
}

func NewDeleteAttachmentUseCase(repo domainAttachment.Repository) *DeleteAttachmentUseCase {
	return &DeleteAttachmentUseCase{repo: repo}
}

func (uc *DeleteAttachmentUseCase) Execute(ctx context.Context, userID, id string) error {
	a, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if a.UserID != userID {
		return domainAttachment.ErrAttachmentNotFound
	}

	return uc.repo.Delete(ctx, id)
}
