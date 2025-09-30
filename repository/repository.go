package repository

import (
	"time"

	"github.com/coocood/freecache"
	"gorm.io/gorm"
)

type Repository struct {
	cache    *freecache.Cache
	database *gorm.DB
	now      func() time.Time
}

const megabyte = 1024 * 1024

func UTCNow() time.Time {
	return time.Now().UTC()
}

func New(database *gorm.DB) *Repository {
	return &Repository{
		database: database,
		cache:    freecache.NewCache(megabyte),
		now:      UTCNow,
	}
}

func (r *Repository) NewTransaction(callback func(tx *gorm.DB) error) error {
	return r.database.Transaction(func(tx *gorm.DB) error {
		return callback(tx)
	})
}
