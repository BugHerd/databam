package databam

import (
	"errors"
)

var (
	ErrNotAPointer      = errors.New("not a pointer")
	ErrNotMappable      = errors.New("can't map to this")
	ErrFieldUnmapped    = errors.New("couldn't map field")
	ErrIncompatibleType = errors.New("incompatible type")
)
