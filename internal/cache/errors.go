package cache

import "errors"

var (
	ErrFailedToGetConnFromPool  = errors.New("failed to get connection from pool")
	ErrFailedToPerformDoCommand = errors.New("failed to perform Do redigo command")
)
