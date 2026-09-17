package cache

import "context"

func FetchOrCache[T any](
	ctx context.Context,
	c Cache,
	key string,
	fetchFromDB func() (T, error),
) (T, error) {
	var result T

	err := c.Get(ctx, key, &result)
	if err == nil {
		return result, nil
	}

	result, err = fetchFromDB()
	if err != nil {
		var zero T
		return zero, err
	}

	_ = c.Set(ctx, key, result, nil)
	return result, nil
}
