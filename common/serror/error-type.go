package serror

import "errors"

var (
	ErrNegativeAmount        = errors.New("amount negative")
	ErrInsufficientAvailable = errors.New("insufficient available")
	ErrInsufficientFrozen    = errors.New("insufficient frozen")
	ErrUnknownAsset          = errors.New("unknown/inactive asset")
)
