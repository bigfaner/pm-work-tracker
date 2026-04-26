package repo

import (
	"database/sql"

	"gorm.io/gorm"
)

// DBTransactor abstracts the gorm.DB.Transaction method for testability.
type DBTransactor interface {
	Transaction(fc func(tx *gorm.DB) error, opts ...*sql.TxOptions) error
}
