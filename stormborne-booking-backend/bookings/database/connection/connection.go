package connection

import (
	"fmt"
	"skyfox/bookings/database/common"
	"skyfox/config"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var once sync.Once

type DBHandler interface {
	Instance() *common.BaseDB
}

type dbHandler struct {
	config config.DbConfig
}

func (dh *dbHandler) Instance() *common.BaseDB {

	var db *gorm.DB
	var err error

	once.Do(func() {
		dsn := connectionString(dh.config)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			//	todo - use the zap logger
		})
	})
	if err != nil {
		panic("could not establish database connection")
	}
	return common.NewBaseDB(db)
}

func NewDBHandler(config config.DbConfig) *dbHandler {
	return &dbHandler{
		config: config,
	}
}

func connectionString(cfg config.DbConfig) string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d", cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port)
}
