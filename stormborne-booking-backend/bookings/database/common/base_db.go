package common

import (
	"context"
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type BaseDB struct {
	db *gorm.DB
}

func NewBaseDB(db *gorm.DB) *BaseDB {
	return &BaseDB{db: db}
}

func (b BaseDB) WithContext(ctx context.Context) (*gorm.DB, func()) {
	context, cancel := context.WithTimeout(ctx, time.Duration(5)*time.Second)
	return b.db.WithContext(context), cancel
}

func (b *BaseDB) SqlDB() (*sql.DB, error) {
	sqlDB, err := b.db.DB()
	if err != nil {
		return nil, err
	}
	return sqlDB, nil
}

func (b *BaseDB) GormDB() *gorm.DB {
	return b.db
}
