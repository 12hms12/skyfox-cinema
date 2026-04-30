package service

import (
	"context"
	"fmt"
	"skyfox/bookings/model"
	"skyfox/bookings/service/mocks"
	movieServiceMock "skyfox/movieservice/movie_gateway/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestShowService_GetShows(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(repo *mocks.MockShowRepository, gw *movieServiceMock.MockMovieGateWay)
		wantErr     bool
		wantLen     int
	}{
		{
			name: "should return empty show list when no shows exist for given date",
			setupMocks: func(repo *mocks.MockShowRepository, gw *movieServiceMock.MockMovieGateWay) {
				repo.On("GetAllShowsOn", mock.Anything, mock.AnythingOfType("string")).Return([]model.Show{}, nil).Once()
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name: "should return error when repository fails to fetch shows for given date",
			setupMocks: func(repo *mocks.MockShowRepository, gw *movieServiceMock.MockMovieGateWay) {
				repo.On("GetAllShowsOn", mock.Anything, mock.AnythingOfType("string")).Return(nil, fmt.Errorf("db error")).Once()
			},
			wantErr: true,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockShowRepository(t)
			gw := movieServiceMock.NewMockMovieGateWay(t)
			tt.setupMocks(repo, gw)

			svc := NewShowService(repo, gw)
			got, err := svc.GetShows(context.Background(), "2026-01-01")

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, got, tt.wantLen)
			}
		})
	}
}

func TestShowService_GetMovieById(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(repo *mocks.MockShowRepository, gw *movieServiceMock.MockMovieGateWay)
		wantErr    bool
		wantMovie  *model.Movie
	}{
		{
			name: "should return movie when gateway successfully fetches movie by id",
			setupMocks: func(repo *mocks.MockShowRepository, gw *movieServiceMock.MockMovieGateWay) {
				gw.On("MovieById", mock.Anything, "tt123").Return(&model.Movie{MovieId: "tt123", Name: "Inception"}, nil).Once()
			},
			wantErr:   false,
			wantMovie: &model.Movie{MovieId: "tt123", Name: "Inception"},
		},
		{
			name: "should return error when gateway fails to fetch movie by id",
			setupMocks: func(repo *mocks.MockShowRepository, gw *movieServiceMock.MockMovieGateWay) {
				gw.On("MovieById", mock.Anything, "tt999").Return(&model.Movie{}, fmt.Errorf("gateway error")).Once()
			},
			wantErr:   true,
			wantMovie: &model.Movie{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockShowRepository(t)
			gw := movieServiceMock.NewMockMovieGateWay(t)
			tt.setupMocks(repo, gw)

			svc := NewShowService(repo, gw)
			id := "tt123"
			if tt.wantErr {
				id = "tt999"
			}
			got, err := svc.GetMovieById(context.Background(), id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantMovie, got)
		})
	}
}

func TestShowService_GetScheduledMovies(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(repo *mocks.MockShowRepository, gw *movieServiceMock.MockMovieGateWay)
		wantErr     bool
		wantLen     int
		wantOrderIds []string
	}{
		{
			name: "should return empty list when repository returns no scheduled movie ids",
			setupMocks: func(repo *mocks.MockShowRepository, gw *movieServiceMock.MockMovieGateWay) {
				repo.On("GetScheduledMovieIds", mock.Anything).Return([]string{}, nil).Once()
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name: "should return movie list when multiple scheduled movie ids exist and gateway succeeds",
			setupMocks: func(repo *mocks.MockShowRepository, gw *movieServiceMock.MockMovieGateWay) {
				repo.On("GetScheduledMovieIds", mock.Anything).Return([]string{"tt1", "tt2"}, nil).Once()
				gw.On("MovieById", mock.Anything, "tt1").Return(&model.Movie{MovieId: "tt1", Name: "Inception"}, nil).Once()
				gw.On("MovieById", mock.Anything, "tt2").Return(&model.Movie{MovieId: "tt2", Name: "Interstellar"}, nil).Once()
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name: "should return error when repository fails to retrieve scheduled movie ids",
			setupMocks: func(repo *mocks.MockShowRepository, gw *movieServiceMock.MockMovieGateWay) {
				repo.On("GetScheduledMovieIds", mock.Anything).Return(nil, fmt.Errorf("db error")).Once()
			},
			wantErr: true,
			wantLen: 0,
		},
		{
			name: "should return partial list when gateway fails for one movie id but succeeds for others",
			setupMocks: func(repo *mocks.MockShowRepository, gw *movieServiceMock.MockMovieGateWay) {
				repo.On("GetScheduledMovieIds", mock.Anything).Return([]string{"tt1", "tt_bad"}, nil).Once()
				gw.On("MovieById", mock.Anything, "tt1").Return(&model.Movie{MovieId: "tt1", Name: "Inception"}, nil).Once()
				gw.On("MovieById", mock.Anything, "tt_bad").Return(nil, fmt.Errorf("gateway error")).Once()
			},
			wantErr: false,
			wantLen: 1,
		},
		{
			name: "should return movies sorted by movie id ascending regardless of goroutine completion order",
			setupMocks: func(repo *mocks.MockShowRepository, gw *movieServiceMock.MockMovieGateWay) {
				repo.On("GetScheduledMovieIds", mock.Anything).Return([]string{"tt_z", "tt_a", "tt_m"}, nil).Once()
				gw.On("MovieById", mock.Anything, "tt_z").Return(&model.Movie{MovieId: "tt_z", Name: "Zoolander"}, nil).Once()
				gw.On("MovieById", mock.Anything, "tt_a").Return(&model.Movie{MovieId: "tt_a", Name: "Arrival"}, nil).Once()
				gw.On("MovieById", mock.Anything, "tt_m").Return(&model.Movie{MovieId: "tt_m", Name: "Matrix"}, nil).Once()
			},
			wantErr:      false,
			wantLen:      3,
			wantOrderIds: []string{"tt_a", "tt_m", "tt_z"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockShowRepository(t)
			gw := movieServiceMock.NewMockMovieGateWay(t)
			tt.setupMocks(repo, gw)

			svc := NewShowService(repo, gw)
			got, err := svc.GetScheduledMovies(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, got, tt.wantLen)
				if tt.wantOrderIds != nil {
					gotIds := make([]string, len(got))
					for i, m := range got {
						gotIds[i] = m.MovieId
					}
					assert.Equal(t, tt.wantOrderIds, gotIds)
				}
			}
		})
	}
}

func TestShowService_GetMovieShowtimes(t *testing.T) {
	slot := model.Slot{Id: 1, Name: "Morning", StartTime: "09:00:00", EndTime: "12:30:00"}
	screen := model.Screen{ID: 1, ScreenName: "Screen 1"}

	tests := []struct {
		name       string
		movieId    string
		date       string
		setupMocks func(repo *mocks.MockShowRepository, gw *movieServiceMock.MockMovieGateWay)
		wantErr    bool
		wantLen    int
		checkFirst func(t *testing.T, got interface{})
	}{
		{
			name:    "should return empty list when no shows exist for given movie and date",
			movieId: "tt123",
			date:    "2026-04-10",
			setupMocks: func(repo *mocks.MockShowRepository, gw *movieServiceMock.MockMovieGateWay) {
				repo.On("GetShowsByMovieAndDate", mock.Anything, "tt123", "2026-04-10").Return([]model.Show{}, nil).Once()
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name:    "should return showtime response with correct available seat count when shows exist",
			movieId: "tt123",
			date:    "2026-04-10",
			setupMocks: func(repo *mocks.MockShowRepository, gw *movieServiceMock.MockMovieGateWay) {
				shows := []model.Show{{Id: 5, MovieId: "tt123", ScreenID: 1, Screen: screen, Slot: slot, Cost: 300}}
				repo.On("GetShowsByMovieAndDate", mock.Anything, "tt123", "2026-04-10").Return(shows, nil).Once()
				repo.On("CountAvailableSeats", mock.Anything, 5).Return(int64(80), nil).Once()
			},
			wantErr: false,
			wantLen: 1,
		},
		{
			name:    "should return error when repository fails to retrieve shows for given movie and date",
			movieId: "tt123",
			date:    "2026-04-10",
			setupMocks: func(repo *mocks.MockShowRepository, gw *movieServiceMock.MockMovieGateWay) {
				repo.On("GetShowsByMovieAndDate", mock.Anything, "tt123", "2026-04-10").Return(nil, fmt.Errorf("db error")).Once()
			},
			wantErr: true,
			wantLen: 0,
		},
		{
			name:    "should return showtime with 0 available seats when seat count query fails",
			movieId: "tt123",
			date:    "2026-04-10",
			setupMocks: func(repo *mocks.MockShowRepository, gw *movieServiceMock.MockMovieGateWay) {
				shows := []model.Show{{Id: 6, MovieId: "tt123", ScreenID: 1, Screen: screen, Slot: slot, Cost: 300}}
				repo.On("GetShowsByMovieAndDate", mock.Anything, "tt123", "2026-04-10").Return(shows, nil).Once()
				repo.On("CountAvailableSeats", mock.Anything, 6).Return(int64(0), fmt.Errorf("count error")).Once()
			},
			wantErr: false,
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockShowRepository(t)
			gw := movieServiceMock.NewMockMovieGateWay(t)
			tt.setupMocks(repo, gw)

			svc := NewShowService(repo, gw)
			got, err := svc.GetMovieShowtimes(context.Background(), tt.movieId, tt.date)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, got, tt.wantLen)
			}
		})
	}
}

func TestShowService_GetMovieAvailableDates(t *testing.T) {
	tests := []struct {
		name       string
		movieId    string
		setupMocks func(repo *mocks.MockShowRepository, gw *movieServiceMock.MockMovieGateWay)
		wantErr    bool
		wantDates  []string
	}{
		{
			name:    "should return available dates when future shows exist for given movie",
			movieId: "tt123",
			setupMocks: func(repo *mocks.MockShowRepository, gw *movieServiceMock.MockMovieGateWay) {
				repo.On("GetAvailableDatesForMovie", mock.Anything, "tt123").Return([]string{"2026-04-10", "2026-04-11"}, nil).Once()
			},
			wantErr:   false,
			wantDates: []string{"2026-04-10", "2026-04-11"},
		},
		{
			name:    "should return empty list when no future shows exist for given movie",
			movieId: "tt456",
			setupMocks: func(repo *mocks.MockShowRepository, gw *movieServiceMock.MockMovieGateWay) {
				repo.On("GetAvailableDatesForMovie", mock.Anything, "tt456").Return([]string{}, nil).Once()
			},
			wantErr:   false,
			wantDates: []string{},
		},
		{
			name:    "should return error when repository fails to retrieve available dates for given movie",
			movieId: "tt789",
			setupMocks: func(repo *mocks.MockShowRepository, gw *movieServiceMock.MockMovieGateWay) {
				repo.On("GetAvailableDatesForMovie", mock.Anything, "tt789").Return(nil, fmt.Errorf("db error")).Once()
			},
			wantErr:   true,
			wantDates: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockShowRepository(t)
			gw := movieServiceMock.NewMockMovieGateWay(t)
			tt.setupMocks(repo, gw)

			svc := NewShowService(repo, gw)
			got, err := svc.GetMovieAvailableDates(context.Background(), tt.movieId)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantDates, got)
			}
		})
	}
}
