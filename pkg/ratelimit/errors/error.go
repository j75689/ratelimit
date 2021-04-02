package errors

import "errors"

// errors of ratelimit
var (
	ErrNotEnoughToken = errors.New("not enough tokens")
)
