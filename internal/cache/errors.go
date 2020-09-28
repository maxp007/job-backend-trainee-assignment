package cache

import (
	"errors"
	"fmt"
)

var (
	ErrContextDeadlineExceeded  = fmt.Errorf("context deadline exceeded")
	ErrFailedToGetConnFromPool  = errors.New("failed to get connection from pool")
	ErrFailedToPerformDoCommand = errors.New("failed to perform Do redigo command")
	ErrKeyLookUpTimeout         = errors.New("failed to lookup key, context timeout")
	ErrKeySetTimeout            = errors.New("failed to set key, context timeout")
)
