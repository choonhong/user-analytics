package domain

import "errors"

var (
	ErrInvalidDate  = errors.New("invalid date format, expected YYYY-MM-DD")
	ErrInvalidMonth = errors.New("invalid month format, expected YYYY-MM")
)
