package db

import (
	"skyfox/bookings/database/common"
	"skyfox/bookings/database/connection"
	"skyfox/bookings/model"
	"skyfox/common/logger"
	"skyfox/config"

	"sync"
)

type testDB struct {
	db *common.BaseDB
}

var once sync.Once
var instance *testDB

func InitDB(cfg config.DbConfig) *testDB {
	once.Do(func() {
		handler := connection.NewDBHandler(cfg)
		db := handler.Instance()
		instance = &testDB{db: db}
	})
	return instance
}

func GetDB() *common.BaseDB {
	return instance.db
}

func (s *testDB) Seed() {

	err := s.db.GormDB().AutoMigrate(
		model.Show{},
		model.Slot{},
		model.Customer{},
		model.Booking{},
		model.Seat{},
		model.ShowSeatStatus{},
		model.BookedSeat{},
		model.ShowPricing{},
		model.Screen{},
		model.ProfilePicture{},
	)
	if err != nil {
		logger.Error("error occurred while migrating schema. error: %v", err)
		return
	}
}
