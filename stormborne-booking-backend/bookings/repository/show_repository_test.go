package repository_test

import (
	"context"
	"skyfox/bookings/model"
	"skyfox/bookings/repository"
	"skyfox/integration_test/db"
	"skyfox/movieservice/movie_gateway/mocks"
	testdata "skyfox/test_data"
	"testing"
	"time"

	repoMocks "skyfox/bookings/repository/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestShowRepository(t *testing.T) {
	//container and database
	db := db.GetDB()

	//migrate
	err := db.GormDB().AutoMigrate(model.Show{}, model.Slot{})
	db.GormDB().Exec(`
		TRUNCATE TABLE booking, show, slot, screen
		RESTART IDENTITY CASCADE
	`)

	db.GormDB().Create(&model.Screen{
		ScreenName: "test",
	})
	db.GormDB().Create(testdata.DummyShows)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	//tests
	t.Run("FindShowById", func(t *testing.T) {
		movieGateway := mocks.NewMockMovieGateWay(t)
		mockSeatRepo := repoMocks.NewMockSeatRepository(t)
		repo := repository.NewShowRepository(db, movieGateway, mockSeatRepo)
		expected := model.Show{
			Id:       1,
			MovieId:  "tt6857189",
			ScreenID: 1,
			Date:     time.Date(2022, time.October, 13, 0, 0, 0, 0, time.UTC),
			SlotId:   3,
			Slot: model.Slot{
				Id:        3,
				Name:      "slot3",
				StartTime: "18:00:00",
				EndTime:   "21:30:00",
			},
			Cost: 300.00,
		}

		actual, _ := repo.FindById(ctx, 1)

		assert.Equal(t, expected, actual)
	})

	t.Run("GetAllShowsByDate", func(t *testing.T) {
		movieGateway := mocks.NewMockMovieGateWay(t)
		mockSeatRepo := repoMocks.NewMockSeatRepository(t)
		repo := repository.NewShowRepository(db, movieGateway, mockSeatRepo)
		expected := []model.Show{
			{
				Id:       1,
				MovieId:  "tt6857189",
				Date:     time.Date(2022, time.October, 13, 0, 0, 0, 0, time.UTC),
				SlotId:   3,
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
				Id:       2,
				MovieId:  "tt6856489",
				Date:     time.Date(2022, time.October, 13, 0, 0, 0, 0, time.UTC),
				SlotId:   4,
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
				Id:       3,
				MovieId:  "tt6856999",
				Date:     time.Date(2022, time.October, 13, 0, 0, 0, 0, time.UTC),
				SlotId:   1,
				ScreenID: 1,
				Slot: model.Slot{
					Id:        1,
					Name:      "slot1",
					StartTime: "09:00:00",
					EndTime:   "12:30:00",
				},
				Cost: 350.00,
			},
		}

		actual, err := repo.GetAllShowsOn(ctx, "2022-10-13")

		assert.Nil(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("Should return correct shows list when valid movieId is passed and valid date", func(t *testing.T) {
		movieGateway := mocks.NewMockMovieGateWay(t)
		mockSeatRepo := repoMocks.NewMockSeatRepository(t)
		repo := repository.NewShowRepository(db, movieGateway, mockSeatRepo)

		movieGateway.On("MovieById", ctx, "tt6856999").Return(&model.Movie{
			Genre: "Drama, Romance",
		}, nil)

		expected := []model.Show{
			{
				Id:       3,
				MovieId:  "tt6856999",
				Date:     time.Date(2022, time.October, 13, 0, 0, 0, 0, time.UTC),
				SlotId:   1,
				ScreenID: 1,
				Slot:     model.Slot{Id: 1, Name: "slot1", StartTime: "09:00:00", EndTime: "12:30:00"},
				Movie: model.Movie{
					Genre: "Drama, Romance",
				},
				Cost: 350.00,
			},
		}

		actual, err := repo.GetAllShowsBy(ctx, &repository.RevenueQuery{
			MovieId:   "tt6856999",
			StartDate: TimePtr(time.Date(2022, time.October, 13, 0, 0, 0, 0, time.UTC)),
			EndDate:   TimePtr(time.Date(2022, time.October, 13, 0, 0, 0, 0, time.UTC)),
		})

		assert.Nil(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("Should return correct shows list when valid movieId is passed and no date", func(t *testing.T) {
		movieGateway := mocks.NewMockMovieGateWay(t)
		mockSeatRepo := repoMocks.NewMockSeatRepository(t)
		repo := repository.NewShowRepository(db, movieGateway, mockSeatRepo)
		movieGateway.On("MovieById", ctx, "tt6856999").Return(&model.Movie{
			Genre: "Drama, Romance",
		}, nil)
		expected := []model.Show{
			{
				Id:       3,
				MovieId:  "tt6856999",
				Date:     time.Date(2022, time.October, 13, 0, 0, 0, 0, time.UTC),
				SlotId:   1,
				ScreenID: 1,
				Slot:     model.Slot{Id: 1, Name: "slot1", StartTime: "09:00:00", EndTime: "12:30:00"},
				Movie: model.Movie{
					Genre: "Drama, Romance",
				},
				Cost: 350.00,
			},
			{
				Id:       4,
				MovieId:  "tt6856999",
				Date:     time.Date(2022, time.October, 14, 0, 0, 0, 0, time.UTC),
				SlotId:   1,
				ScreenID: 1,
				Slot:     model.Slot{Id: 1, Name: "slot1", StartTime: "09:00:00", EndTime: "12:30:00"},
				Movie: model.Movie{
					Genre: "Drama, Romance",
				},
				Cost: 330.00,
			},
		}

		actual, err := repo.GetAllShowsBy(ctx, &repository.RevenueQuery{
			MovieId: "tt6856999",
		})

		assert.Nil(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("Should return correct shows list when valid date is passed and no movieId", func(t *testing.T) {
		movieGateway := mocks.NewMockMovieGateWay(t)
		mockSeatRepo := repoMocks.NewMockSeatRepository(t)
		repo := repository.NewShowRepository(db, movieGateway, mockSeatRepo)
		movieGateway.On("MovieById", ctx, mock.AnythingOfType("string")).Return(&model.Movie{
			Genre: "Drama, Romance",
		}, nil)
		expected := []model.Show{
			{
				Id:       1,
				MovieId:  "tt6857189",
				Date:     time.Date(2022, time.October, 13, 0, 0, 0, 0, time.UTC),
				SlotId:   3,
				ScreenID: 1,
				Slot:     model.Slot{Id: 3, Name: "slot3", StartTime: "18:00:00", EndTime: "21:30:00"},
				Movie: model.Movie{
					Genre: "Drama, Romance",
				},
				Cost: 300.00,
			},
			{
				Id:       2,
				MovieId:  "tt6856489",
				Date:     time.Date(2022, time.October, 13, 0, 0, 0, 0, time.UTC),
				SlotId:   4,
				ScreenID: 1,
				Slot:     model.Slot{Id: 4, Name: "slot4", StartTime: "22:30:00", EndTime: "02:00:00"},
				Movie: model.Movie{
					Genre: "Drama, Romance",
				},
				Cost: 350.00,
			},
			{
				Id:       3,
				MovieId:  "tt6856999",
				Date:     time.Date(2022, time.October, 13, 0, 0, 0, 0, time.UTC),
				SlotId:   1,
				ScreenID: 1,
				Slot:     model.Slot{Id: 1, Name: "slot1", StartTime: "09:00:00", EndTime: "12:30:00"},
				Movie: model.Movie{
					Genre: "Drama, Romance",
				},
				Cost: 350.00,
			},
		}

		actual, err := repo.GetAllShowsBy(ctx, &repository.RevenueQuery{
			MovieId:   "",
			StartDate: TimePtr(time.Date(2022, time.October, 13, 0, 0, 0, 0, time.UTC)),
			EndDate:   TimePtr(time.Date(2022, time.October, 13, 0, 0, 0, 0, time.UTC)),
		})

		assert.Nil(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("Should return all shows when no date and no movieId", func(t *testing.T) {
		movieGateway := mocks.NewMockMovieGateWay(t)
		mockSeatRepo := repoMocks.NewMockSeatRepository(t)
		repo := repository.NewShowRepository(db, movieGateway, mockSeatRepo)
		movieGateway.On("MovieById", ctx, mock.AnythingOfType("string")).Return(&model.Movie{
			Genre: "Drama, Romance",
		}, nil)
		expected := []model.Show{
			{
				Id:       1,
				MovieId:  "tt6857189",
				Date:     time.Date(2022, time.October, 13, 0, 0, 0, 0, time.UTC),
				SlotId:   3,
				ScreenID: 1,
				Slot:     model.Slot{Id: 3, Name: "slot3", StartTime: "18:00:00", EndTime: "21:30:00"},
				Movie: model.Movie{
					Genre: "Drama, Romance",
				},
				Cost: 300.00,
			},
			{
				Id:       2,
				MovieId:  "tt6856489",
				Date:     time.Date(2022, time.October, 13, 0, 0, 0, 0, time.UTC),
				SlotId:   4,
				ScreenID: 1,
				Slot:     model.Slot{Id: 4, Name: "slot4", StartTime: "22:30:00", EndTime: "02:00:00"},
				Movie: model.Movie{
					Genre: "Drama, Romance",
				},
				Cost: 350.00,
			},
			{
				Id:       3,
				MovieId:  "tt6856999",
				Date:     time.Date(2022, time.October, 13, 0, 0, 0, 0, time.UTC),
				SlotId:   1,
				ScreenID: 1,
				Movie: model.Movie{
					Genre: "Drama, Romance",
				},
				Slot: model.Slot{Id: 1, Name: "slot1", StartTime: "09:00:00", EndTime: "12:30:00"},
				Cost: 350.00,
			},
			{
				Id:       4,
				MovieId:  "tt6856999",
				Date:     time.Date(2022, time.October, 14, 0, 0, 0, 0, time.UTC),
				SlotId:   1,
				ScreenID: 1,
				Slot:     model.Slot{Id: 1, Name: "slot1", StartTime: "09:00:00", EndTime: "12:30:00"},
				Movie: model.Movie{
					Genre: "Drama, Romance",
				},
				Cost: 330.00,
			},
		}

		actual, err := repo.GetAllShowsBy(ctx, &repository.RevenueQuery{})

		assert.Nil(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("Should filter shows by genre correctly", func(t *testing.T) {
		movieGateway := mocks.NewMockMovieGateWay(t)
		mockSeatRepo := repoMocks.NewMockSeatRepository(t)
		repo := repository.NewShowRepository(db, movieGateway, mockSeatRepo)

		movieGateway.On("MovieById", ctx, "tt6856999").Return(&model.Movie{
			Genre: "Drama, Romance",
		}, nil)
		movieGateway.On("MovieById", ctx, mock.Anything).Return(&model.Movie{
			Genre: "Action, Sci-Fi",
		}, nil)

		query := &repository.RevenueQuery{
			Genre: "Drama",
		}

		actual, err := repo.GetAllShowsBy(ctx, query)

		assert.Nil(t, err)
		assert.Len(t, actual, 2)
		for _, show := range actual {
			assert.Equal(t, "tt6856999", show.MovieId)
		}
	})

	t.Run("Should return empty list when movieId doesn't exist", func(t *testing.T) {
		movieGateway := mocks.NewMockMovieGateWay(t)
		mockSeatRepo := repoMocks.NewMockSeatRepository(t)
		repo := repository.NewShowRepository(db, movieGateway, mockSeatRepo)

		actual, err := repo.GetAllShowsBy(ctx, &repository.RevenueQuery{
			MovieId: "1234",
		})

		assert.NoError(t, err)
		assert.Empty(t, actual)
	})

	t.Run("Should return empty list when start date is after end date", func(t *testing.T) {
		movieGateway := mocks.NewMockMovieGateWay(t)
		mockSeatRepo := repoMocks.NewMockSeatRepository(t)
		repo := repository.NewShowRepository(db, movieGateway, mockSeatRepo)
		actual, err := repo.GetAllShowsBy(ctx, &repository.RevenueQuery{
			MovieId:   "1234",
			StartDate: TimePtr(time.Date(2022, time.October, 13, 0, 0, 0, 0, time.UTC)),
			EndDate:   TimePtr(time.Date(2021, time.October, 13, 0, 0, 0, 0, time.UTC)),
		})

		assert.NoError(t, err)
		assert.Empty(t, actual)
	})

}

func TestShowRepository_GetShowsByMovieAndDate(t *testing.T) {
	db := db.GetDB()
	db.GormDB().AutoMigrate(model.Show{}, model.Slot{})
	db.GormDB().Exec(`TRUNCATE TABLE booking, show, slot, screen RESTART IDENTITY CASCADE`)
	db.GormDB().Create(&model.Screen{ScreenName: "test"})
	db.GormDB().Create(testdata.DummyShows)

	ctx := context.Background()

	tests := []struct {
		name     string
		movieId  string
		date     string
		wantLen  int
		wantErr  bool
	}{
		{
			name:    "should return shows matching movieId and date when valid inputs are provided",
			movieId: "tt6856999",
			date:    "2022-10-13",
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "should return empty list when movieId does not match any show on given date",
			movieId: "tt9999999",
			date:    "2022-10-13",
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "should return empty list when no shows exist for movieId on given date",
			movieId: "tt6857189",
			date:    "2022-10-14",
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "should return error when invalid date format is provided",
			movieId: "tt6856999",
			date:    "not-a-date",
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			movieGateway := mocks.NewMockMovieGateWay(t)
			mockSeatRepo := repoMocks.NewMockSeatRepository(t)
			repo := repository.NewShowRepository(db, movieGateway, mockSeatRepo)

			got, err := repo.GetShowsByMovieAndDate(ctx, tt.movieId, tt.date)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, got, tt.wantLen)
			}
		})
	}
}

func TestShowRepository_CountAvailableSeats(t *testing.T) {
	db := db.GetDB()
	db.GormDB().AutoMigrate(model.Show{}, model.Slot{}, model.ShowSeatStatus{})
	db.GormDB().Exec(`TRUNCATE TABLE booking, show_seat_status, show, slot, screen RESTART IDENTITY CASCADE`)
	db.GormDB().Create(&model.Screen{ScreenName: "count-test"})
	db.GormDB().Create(testdata.DummyShows)

	ctx := context.Background()

	tests := []struct {
		name      string
		setupFunc func(repo interface{ CreateShow(context.Context, *model.Show) error }) int
		wantCount int64
	}{
		{
			name: "should return 0 when no seat status entries exist for given show",
			setupFunc: func(_ interface{ CreateShow(context.Context, *model.Show) error }) int {
				return 1
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			movieGateway := mocks.NewMockMovieGateWay(t)
			mockSeatRepo := repoMocks.NewMockSeatRepository(t)
			repo := repository.NewShowRepository(db, movieGateway, mockSeatRepo)

			showId := tt.setupFunc(repo)
			count, err := repo.CountAvailableSeats(ctx, showId)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantCount, count)
		})
	}
}

func TestShowRepository_GetScheduledMovieIds(t *testing.T) {
	db := db.GetDB()
	db.GormDB().Exec(`TRUNCATE TABLE booking, show, slot, screen RESTART IDENTITY CASCADE`)
	db.GormDB().Create(&model.Screen{ScreenName: "sched-test"})
	db.GormDB().Exec(`INSERT INTO slot (name, start_time, end_time) VALUES ('slot1','09:00:00','12:30:00'),('slot2','13:30:00','17:00:00'),('slot3','18:00:00','21:30:00'),('slot4','22:30:00','02:00:00')`)

	ctx := context.Background()

	tests := []struct {
		name      string
		extraData []model.Show
		wantEmpty bool
		wantIds   []string
	}{
		{
			name:      "should return empty list when show table contains no future-dated shows",
			extraData: nil,
			wantEmpty: true,
		},
		{
			name: "should return distinct movie ids sorted ascending when future shows exist",
			extraData: []model.Show{
				{MovieId: "tt_future_2", Date: time.Now().UTC().Add(48 * time.Hour), SlotId: 1, ScreenID: 1, Cost: 250},
				{MovieId: "tt_future_1", Date: time.Now().UTC().Add(48 * time.Hour), SlotId: 3, ScreenID: 1, Cost: 200},
				{MovieId: "tt_future_1", Date: time.Now().UTC().Add(72 * time.Hour), SlotId: 4, ScreenID: 1, Cost: 200},
			},
			wantEmpty: false,
			wantIds:   []string{"tt_future_1", "tt_future_2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			movieGateway := mocks.NewMockMovieGateWay(t)
			mockSeatRepo := repoMocks.NewMockSeatRepository(t)
			repo := repository.NewShowRepository(db, movieGateway, mockSeatRepo)

			var inserted []model.Show
			if len(tt.extraData) > 0 {
				for i := range tt.extraData {
					db.GormDB().Create(&tt.extraData[i])
					inserted = append(inserted, tt.extraData[i])
				}
			}
			defer func() {
				for _, s := range inserted {
					db.GormDB().Delete(&model.Show{}, s.Id)
				}
			}()

			got, err := repo.GetScheduledMovieIds(ctx)

			assert.NoError(t, err)
			if tt.wantEmpty {
				assert.Empty(t, got)
			} else {
				assert.Equal(t, tt.wantIds, got)
			}
		})
	}
}

func TestShowRepository_GetAvailableDatesForMovie(t *testing.T) {
	db := db.GetDB()
	db.GormDB().Exec(`TRUNCATE TABLE booking, show, slot, screen RESTART IDENTITY CASCADE`)
	db.GormDB().Create(&model.Screen{ScreenName: "dates-test"})
	db.GormDB().Exec(`INSERT INTO slot (name, start_time, end_time) VALUES ('slot1','09:00:00','12:30:00'),('slot2','13:30:00','17:00:00'),('slot3','18:00:00','21:30:00'),('slot4','22:30:00','02:00:00')`)

	ctx := context.Background()

	tests := []struct {
		name      string
		movieId   string
		extraData []model.Show
		wantEmpty bool
		wantLen   int
	}{
		{
			name:      "should return empty list when no future shows exist for given movieId",
			movieId:   "tt_unknown_movie",
			extraData: nil,
			wantEmpty: true,
		},
		{
			name:    "should return formatted dates when future shows exist for given movieId",
			movieId: "tt_dates_movie",
			extraData: []model.Show{
				{MovieId: "tt_dates_movie", Date: time.Now().UTC().Add(24 * time.Hour), SlotId: 3, ScreenID: 1, Cost: 200},
				{MovieId: "tt_dates_movie", Date: time.Now().UTC().Add(48 * time.Hour), SlotId: 4, ScreenID: 1, Cost: 200},
			},
			wantEmpty: false,
			wantLen:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			movieGateway := mocks.NewMockMovieGateWay(t)
			mockSeatRepo := repoMocks.NewMockSeatRepository(t)
			repo := repository.NewShowRepository(db, movieGateway, mockSeatRepo)

			var inserted []model.Show
			if len(tt.extraData) > 0 {
				for i := range tt.extraData {
					db.GormDB().Create(&tt.extraData[i])
					inserted = append(inserted, tt.extraData[i])
				}
			}
			defer func() {
				for _, s := range inserted {
					db.GormDB().Delete(&model.Show{}, s.Id)
				}
			}()

			got, err := repo.GetAvailableDatesForMovie(ctx, tt.movieId)

			assert.NoError(t, err)
			if tt.wantEmpty {
				assert.Empty(t, got)
			} else {
				assert.Len(t, got, tt.wantLen)
				for _, d := range got {
					assert.Regexp(t, `^\d{4}-\d{2}-\d{2}$`, d)
				}
			}
		})
	}
}

func TimePtr(t time.Time) *time.Time {
	return &t
}
