package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestMapNotFound_GormErrRecordNotFound(t *testing.T) {
	domainErr := ErrItemNotFound
	result := MapNotFound(gorm.ErrRecordNotFound, domainErr)
	assert.Equal(t, domainErr, result)
}

func TestMapNotFound_ErrNotFound(t *testing.T) {
	domainErr := ErrTeamNotFound
	result := MapNotFound(ErrNotFound, domainErr)
	assert.Equal(t, domainErr, result)
}

func TestMapNotFound_OtherError(t *testing.T) {
	otherErr := errors.New("some database error")
	domainErr := ErrItemNotFound
	result := MapNotFound(otherErr, domainErr)
	assert.Equal(t, otherErr, result)
}

func TestMapNotFound_NilError(t *testing.T) {
	domainErr := ErrItemNotFound
	result := MapNotFound(nil, domainErr)
	assert.Nil(t, result)
}

func TestMapNotFound_WrappedGormErr(t *testing.T) {
	wrappedErr := errors.Join(errors.New("context"), gorm.ErrRecordNotFound)
	domainErr := ErrItemNotFound
	result := MapNotFound(wrappedErr, domainErr)
	assert.Equal(t, domainErr, result)
}
