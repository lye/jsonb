package jsonb

import (
	"errors"
)

var (
	ErrNoSuchKind      = errors.New("jsonb: no such kind")
	ErrInvalidScanType = errors.New("jsonb: invalid scan type")
	ErrInvalidJsonType = errors.New("jsonb: invalid json type")
)
