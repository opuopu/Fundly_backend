package attachment

import (
	"context"

	domainAttachment "github.com/Dodul01/fundly/internal/domain/attachment"
)

type GetAttachmentUseCase struct {
	repo domainAttachment.Repository
}

func NewGetAttachmentUseCase(repo domainAttachment.Repository) *GetAttachmentUseCase {
	return &GetAttachmentUseCase{repo: repo}
}

func (uc *GetAttachmentUseCase) Execute(ctx context.Context, userID, id string) (*AttachmentResponse, error) {
	a, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if a.UserID != userID {
		return nil, domainAttachment.ErrAttachmentNotFound
	}

	res := toResponse(a)
	return &res, nil
}
