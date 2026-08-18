package transaction

import (
	"context"
	"ms_category/internal/core/cache"
)

type WriteExecutor struct {
	tx         Manager
	cache      cache.Cache
	keyBuilder cache.KeyBuilder
}

func NewWriterExecutor(
	tx Manager,
	cache cache.Cache,
	keyBuilder cache.KeyBuilder,
) *WriteExecutor {
	return &WriteExecutor{
		tx:         tx,
		cache:      cache,
		keyBuilder: keyBuilder,
	}
}

func (e *WriteExecutor) Execute(
	ctx context.Context,
	fn func(ctx context.Context) error,
) error {
	err := e.tx.RunInTx(ctx, fn)
	if err != nil {
		return err
	}

	go e.invalidateCache(ctx)
	return nil
}

func (e *WriteExecutor) invalidateCache(ctx context.Context) {
	_ = e.cache.DeleteByPrefix(ctx, e.keyBuilder.GetPrefix())
}
