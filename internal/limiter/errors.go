package limiter

import "errors"

var (
	ErrServiceUnavailable = errors.New("rate guard service unavailable")
	ErrInternal           = errors.New("rate guard internal error")
)
