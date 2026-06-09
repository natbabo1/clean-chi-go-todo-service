package todo

import "errors"

var (
	ErrNotFound  = errors.New("todo not found")
	ErrForbidden = errors.New("forbidden")
)
