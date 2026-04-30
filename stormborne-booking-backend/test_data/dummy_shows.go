package testdata

import (
	"skyfox/bookings/model"
	"time"
)

var DummyShows = []model.Show{
	{
		Id:      1,
		MovieId: "tt6857189",
		Date:    time.Date(2022,time.October,13,0,0,0,0,time.UTC),
		SlotId:  3,
		ScreenID: 1,
		Slot: model.Slot{
			Id:        3,
			Name:      "slot3",
			StartTime: "18:00:00",
			EndTime:   "21:30:00",
		},
		Cost: 300.00,
	},
	{
		Id:      2,
		MovieId: "tt6856489",
		Date:    time.Date(2022,time.October,13,0,0,0,0,time.UTC),
		SlotId:  4,
		ScreenID: 1,
		Slot: model.Slot{
			Id:        4,
			Name:      "slot4",
			StartTime: "22:30:00",
			EndTime:   "02:00:00",
		},
		Cost: 350.00,
	},
	{
		Id:      3,
		MovieId: "tt6856999",
		Date:    time.Date(2022,time.October,13,0,0,0,0,time.UTC),
		SlotId:  1,
		ScreenID: 1,
		Slot: model.Slot{
			Id:        1,
			Name:      "slot1",
			StartTime: "09:00:00",
			EndTime:   "12:30:00",
		},
		Cost: 350.00,
	},
	{
		Id:      4,
		MovieId: "tt6856999",
		ScreenID: 1,
		Date:    time.Date(2022,time.October,14,0,0,0,0,time.UTC),
		SlotId:  1,
		Slot: model.Slot{
			Id:        1,
			Name:      "slot1",
			StartTime: "09:00:00",
			EndTime:   "12:30:00",
		},
		Cost: 330.00,
	},
}
