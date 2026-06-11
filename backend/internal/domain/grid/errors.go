package grid

import "errors"

var (
	ErrGridNotFound  = errors.New("grid not found")
	ErrQueryNotFound = errors.New("query not found")
)
