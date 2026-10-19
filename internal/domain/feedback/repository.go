package feedback

import "context"

type Repository interface {
	Create(ctx context.Context, f *Feedback) error
}
