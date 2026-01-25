package gorm

import (
	"errors"

	"gorm.io/gorm"
)

// ErrRecordNotFound ...
var ErrRecordNotFound = gorm.ErrRecordNotFound

// IsRecordNotFoundError returns true if error contains a RecordNotFound error
func IsRecordNotFoundError(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
