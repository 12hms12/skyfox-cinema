package repository_test

import (
	"context"
	"testing"
	"time"

	"skyfox/bookings/database/common"
	"skyfox/bookings/model"
	"skyfox/bookings/repository"
	"skyfox/integration_test/db"

	"github.com/stretchr/testify/assert"
)

type testFixtures struct {
	screenID int
	showID   int
	seatID   int
}

func setupDB(t *testing.T) (repository.SeatRepository, *common.BaseDB) {
	t.Helper()

	database := db.GetDB()
	database.GormDB().AutoMigrate(
		&model.Slot{},
		&model.Screen{},
		&model.Seat{},
		&model.Show{},
		&model.ShowPricing{},
		&model.ShowSeatStatus{},
	)

	database.GormDB().Exec("TRUNCATE show_seat_status CASCADE")
	database.GormDB().Exec("TRUNCATE show_pricing CASCADE")
	database.GormDB().Exec("TRUNCATE show CASCADE")
	database.GormDB().Exec("TRUNCATE seat CASCADE")
	database.GormDB().Exec("TRUNCATE screen CASCADE")
	database.GormDB().Exec("TRUNCATE slot CASCADE")

	database.GormDB().Create(&model.Slot{Name: "Morning", StartTime: "09:00", EndTime: "12:00"})

	return repository.NewSeatRepository(database), database
}


func seedFixtures(t *testing.T, repo repository.SeatRepository, database *common.BaseDB, ctx context.Context) testFixtures {
	t.Helper()

	screen := &model.Screen{ScreenName: "Screen 1"}
	if err := database.GormDB().Create(screen).Error; err != nil {
		t.Fatalf("seed: create screen: %v", err)
	}

	classicSeat := &model.Seat{ScreenID: screen.ID, RowNumber: 1, ColumnNumber: 1, SeatType: "CLASSIC"}
	if err := database.GormDB().Create(classicSeat).Error; err != nil {
		t.Fatalf("seed: create classic seat: %v", err)
	}

	if err := database.GormDB().Create(&model.Seat{
		ScreenID: screen.ID, RowNumber: 6, ColumnNumber: 1, SeatType: "CLUB",
	}).Error; err != nil {
		t.Fatalf("seed: create club seat: %v", err)
	}

	if err := database.GormDB().Create(&model.Seat{
		ScreenID: screen.ID, RowNumber: 9, ColumnNumber: 1, SeatType: "RECLINER",
	}).Error; err != nil {
		t.Fatalf("seed: create recliner seat: %v", err)
	}

	var slot model.Slot
	database.GormDB().First(&slot)

	show := &model.Show{
		ScreenID: screen.ID,
		MovieId:  "MOVIE-001",
		Date:     time.Now().AddDate(0, 0, 1),
		Cost:     200.0,
		SlotId:   slot.Id,
	}
	if err := repo.CreateShow(ctx, show); err != nil {
		t.Fatalf("seed: create show: %v", err)
	}

	for _, p := range []model.ShowPricing{
		{ShowID: show.Id, SeatType: "CLASSIC", Price: 200.0},
		{ShowID: show.Id, SeatType: "CLUB", Price: 250.0},
		{ShowID: show.Id, SeatType: "RECLINER", Price: 300.0},
	} {
		pCopy := p
		if err := repo.CreateShowPricing(ctx, &pCopy); err != nil {
			t.Fatalf("seed: create pricing %s: %v", p.SeatType, err)
		}
	}

	sss := &model.ShowSeatStatus{ShowID: show.Id, SeatID: classicSeat.ID, Status: "AVAILABLE"}
	if err := repo.CreateShowSeatStatus(ctx, sss); err != nil {
		t.Fatalf("seed: create seat status: %v", err)
	}

	return testFixtures{
		screenID: screen.ID,
		showID:   show.Id,
		seatID:   classicSeat.ID,
	}
}

func TestSeatRepository(t *testing.T) {
	repo, database := setupDB(t)
	ctx := context.Background()
	fx := seedFixtures(t, repo, database, ctx)

	t.Run("FindScreenByName", func(t *testing.T) {
		tests := []struct {
			name       string
			screenName string
			wantNil    bool
		}{
			{
				name:       "non-existent screen returns nil",
				screenName: "NonExistent",
				wantNil:    true,
			},
			{
				name:       "existing screen is returned",
				screenName: "Screen 1",
				wantNil:    false,
			},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				screen, err := repo.FindScreenByName(ctx, tc.screenName)
				assert.NoError(t, err)
				if tc.wantNil {
					assert.Nil(t, screen)
				} else {
					assert.NotNil(t, screen)
					assert.Equal(t, tc.screenName, screen.ScreenName)
				}
			})
		}
	})

	t.Run("FindSeatsByScreenID", func(t *testing.T) {
		tests := []struct {
			name      string
			screenID  int
			wantEmpty bool
			wantTypes []string 
		}{
			{
				name:      "returns seats for existing screen with correct types",
				screenID:  fx.screenID,
				wantEmpty: false,
				wantTypes: []string{"CLASSIC", "CLUB", "RECLINER"},
			},
			{
				name:      "returns empty for non-existent screen",
				screenID:  99999,
				wantEmpty: true,
				wantTypes: nil,
			},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				seats, err := repo.FindSeatsByScreenID(ctx, tc.screenID)
				assert.NoError(t, err)
				if tc.wantEmpty {
					assert.Empty(t, seats)
				} else {
					assert.NotEmpty(t, seats)
					actualTypes := make(map[string]bool)
					for _, s := range seats {
						actualTypes[s.SeatType] = true
					}
					for _, wantType := range tc.wantTypes {
						assert.True(t, actualTypes[wantType], "expected seat type %s not found", wantType)
					}
				}
			})
		}
	})

	t.Run("CreateSeat", func(t *testing.T) {
		tests := []struct {
			name     string
			seatType string
			row      int
			col      int
		}{
			{
				name:     "persists a CLASSIC seat",
				seatType: "CLASSIC",
				row:      2,
				col:      1,
			},
			{
				name:     "persists a CLUB seat",
				seatType: "CLUB",
				row:      7,
				col:      1,
			},
			{
				name:     "persists a RECLINER seat",
				seatType: "RECLINER",
				row:      9,
				col:      2,
			},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				seat := &model.Seat{
					ScreenID:     fx.screenID,
					RowNumber:    tc.row,
					ColumnNumber: tc.col,
					SeatType:     tc.seatType,
				}
				err := repo.CreateSeat(ctx, seat)
				assert.NoError(t, err)
				assert.NotZero(t, seat.ID)
			})
		}
	})

	t.Run("FindShowByScreenAndMovie", func(t *testing.T) {
		tests := []struct {
			name     string
			screenID int
			movieID  string
			wantNil  bool
		}{
			{
				name:     "non-existent movie returns nil",
				screenID: fx.screenID,
				movieID:  "MOVIE-NONEXISTENT",
				wantNil:  true,
			},
			{
				name:     "existing show is returned",
				screenID: fx.screenID,
				movieID:  "MOVIE-001",
				wantNil:  false,
			},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				show, err := repo.FindShowByScreenAndMovie(ctx, tc.screenID, tc.movieID)
				assert.NoError(t, err)
				if tc.wantNil {
					assert.Nil(t, show)
				} else {
					assert.NotNil(t, show)
					assert.Equal(t, tc.movieID, show.MovieId)
				}
			})
		}
	})

	t.Run("CreateShow", func(t *testing.T) {
		var slot model.Slot
		database.GormDB().First(&slot)

		tests := []struct {
			name    string
			movieID string
			cost    float64
		}{
			{
				name:    "persists a new show record",
				movieID: "MOVIE-002",
				cost:    150.0,
			},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				show := &model.Show{
					ScreenID: fx.screenID,
					MovieId:  tc.movieID,
					Date:     time.Now().AddDate(0, 0, 2),
					Cost:     tc.cost,
					SlotId:   slot.Id,
				}
				err := repo.CreateShow(ctx, show)
				assert.NoError(t, err)
				assert.NotZero(t, show.Id)
			})
		}
	})

	t.Run("GetShowByID", func(t *testing.T) {
		tests := []struct {
			name    string
			showID  uint
			wantErr bool
		}{
			{
				name:    "valid show ID returns show",
				showID:  uint(fx.showID),
				wantErr: false,
			},
			{
				name:    "non-existent show ID returns error",
				showID:  99999,
				wantErr: true,
			},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				show, err := repo.GetShowByID(ctx, tc.showID)
				if tc.wantErr {
					assert.Error(t, err)
					assert.Nil(t, show)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, show)
					assert.Equal(t, fx.showID, show.Id)
				}
			})
		}
	})

	t.Run("FindPricingByShowID", func(t *testing.T) {
		tests := []struct {
			name      string
			showID    int
			wantLen   int
			wantTypes []string
		}{
			{
				name:      "returns all 3 pricing types for valid show",
				showID:    fx.showID,
				wantLen:   3,
				wantTypes: []string{"CLASSIC", "CLUB", "RECLINER"},
			},
			{
				name:      "returns empty for show with no pricing",
				showID:    99999,
				wantLen:   0,
				wantTypes: nil,
			},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				pricing, err := repo.FindPricingByShowID(ctx, tc.showID)
				assert.NoError(t, err)
				assert.Len(t, pricing, tc.wantLen)
				if tc.wantLen > 0 {
					actualTypes := make(map[string]bool)
					for _, p := range pricing {
						actualTypes[p.SeatType] = true
					}
					for _, wantType := range tc.wantTypes {
						assert.True(t, actualTypes[wantType], "expected pricing type %s not found", wantType)
					}
				}
			})
		}
	})

	t.Run("CreateShowPricing", func(t *testing.T) {
		tests := []struct {
			name     string
			seatType string
			price    float64
		}{
			{
				name:     "persists a CLASSIC pricing record",
				seatType: "CLASSIC",
				price:    200.0,
			},
			{
				name:     "persists a CLUB pricing record",
				seatType: "CLUB",
				price:    250.0,
			},
			{
				name:     "persists a RECLINER pricing record",
				seatType: "RECLINER",
				price:    300.0,
			},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				pricing := &model.ShowPricing{
					ShowID:   fx.showID,
					SeatType: tc.seatType,
					Price:    tc.price,
				}
				err := repo.CreateShowPricing(ctx, pricing)
				assert.NoError(t, err)
				assert.NotZero(t, pricing.ID)
			})
		}
	})

	t.Run("GetPricingByShowID", func(t *testing.T) {
		tests := []struct {
			name      string
			showID    uint
			wantEmpty bool
		}{
			{
				name:      "returns pricing for valid show",
				showID:    uint(fx.showID),
				wantEmpty: false,
			},
			{
				name:      "returns empty for show with no pricing",
				showID:    99999,
				wantEmpty: true,
			},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				pricing, err := repo.GetPricingByShowID(ctx, tc.showID)
				assert.NoError(t, err)
				if tc.wantEmpty {
					assert.Empty(t, pricing)
				} else {
					assert.NotEmpty(t, pricing)
				}
			})
		}
	})

	t.Run("FindSeatStatusByShowID", func(t *testing.T) {
		tests := []struct {
			name      string
			showID    int
			wantLen   int
			wantState string
		}{
			{
				name:      "returns statuses for given show",
				showID:    fx.showID,
				wantLen:   1,
				wantState: "AVAILABLE",
			},
			{
				name:      "returns empty for non-existent show",
				showID:    99999,
				wantLen:   0,
				wantState: "",
			},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				statuses, err := repo.FindSeatStatusByShowID(ctx, tc.showID)
				assert.NoError(t, err)
				assert.Len(t, statuses, tc.wantLen)
				if tc.wantLen > 0 {
					assert.Equal(t, tc.wantState, statuses[0].Status)
				}
			})
		}
	})

	t.Run("CreateShowSeatStatus", func(t *testing.T) {
		tests := []struct {
			name   string
			status string
		}{
			{
				name:   "persists an AVAILABLE seat status",
				status: "AVAILABLE",
			},
			{
				name:   "persists a BOOKED seat status",
				status: "BOOKED",
			},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				sss := &model.ShowSeatStatus{
					ShowID: fx.showID,
					SeatID: fx.seatID,
					Status: tc.status,
				}
				err := repo.CreateShowSeatStatus(ctx, sss)
				assert.NoError(t, err)
				assert.NotZero(t, sss.ID)
			})
		}
	})

	t.Run("GetSeatStatusByShowID", func(t *testing.T) {
		database.GormDB().Exec("DELETE FROM show_seat_status WHERE show_id = ?", fx.showID)

		expiredTime := time.Now().Add(-1 * time.Minute)
		database.GormDB().Create(&model.ShowSeatStatus{ShowID: fx.showID, SeatID: fx.seatID, Status: "AVAILABLE"})
		database.GormDB().Create(&model.ShowSeatStatus{ShowID: fx.showID, SeatID: fx.seatID, Status: "BOOKED"})
		database.GormDB().Create(&model.ShowSeatStatus{ShowID: fx.showID, SeatID: fx.seatID, Status: "BLOCKED", LockedUntil: &expiredTime})

		tests := []struct {
			name        string
			showID      uint
			wantEmpty   bool
			wantPreload bool 
		}{
			{
				name:        "returns all statuses with seat preloaded for valid show",
				showID:      uint(fx.showID),
				wantEmpty:   false,
				wantPreload: true,
			},
			{
				name:        "returns empty for non-existent show",
				showID:      99999,
				wantEmpty:   true,
				wantPreload: false,
			},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				statuses, err := repo.GetSeatStatusByShowID(ctx, tc.showID)
				assert.NoError(t, err)
				if tc.wantEmpty {
					assert.Empty(t, statuses)
				} else {
					assert.NotEmpty(t, statuses)
					assert.NotZero(t, statuses[0].Seat.ID)
				}
			})
		}
	})

	t.Run("FindShowsByScreenID", func(t *testing.T) {
		tests := []struct {
			name      string
			screenID  int
			wantEmpty bool
			wantMovie string
		}{
			{
				name:      "returns all shows for given screen",
				screenID:  fx.screenID,
				wantEmpty: false,
				wantMovie: "MOVIE-001",
			},
			{
				name:      "returns empty for non-existent screen",
				screenID:  99999,
				wantEmpty: true,
				wantMovie: "",
			},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				shows, err := repo.FindShowsByScreenID(ctx, tc.screenID)
				assert.NoError(t, err)
				if tc.wantEmpty {
					assert.Empty(t, shows)
				} else {
					assert.NotEmpty(t, shows)
					assert.Equal(t, tc.wantMovie, shows[0].MovieId)
				}
			})
		}
	})
}