package executor

import "context"

type Executor interface {
	ScaleUp(ctx context.Context, count int) error
	ScaleDown(ctx context.Context, count int) error
	CurrentCapacity(ctx context.Context) (int, error)
	CurrentResourceIDs(ctx context.Context) ([]string, error)
}
