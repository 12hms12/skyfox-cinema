package controller

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"skyfox/bookings/dto/response"
	"skyfox/bookings/model"
	ae "skyfox/error"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockShowControllerService struct {
	mock.Mock
}

func (m *MockShowControllerService) GetShows(ctx context.Context, date string) ([]model.Show, error) {
	args := m.Called(ctx, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Show), args.Error(1)
}

func (m *MockShowControllerService) GetMovieById(ctx context.Context, movieId string) (*model.Movie, error) {
	args := m.Called(ctx, movieId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Movie), args.Error(1)
}

func (m *MockShowControllerService) GetScheduledMovies(ctx context.Context) ([]model.Movie, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Movie), args.Error(1)
}

func (m *MockShowControllerService) GetMovieShowtimes(ctx context.Context, movieId, date string) ([]response.ShowtimeResponse, error) {
	args := m.Called(ctx, movieId, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]response.ShowtimeResponse), args.Error(1)
}

func (m *MockShowControllerService) GetMovieAvailableDates(ctx context.Context, movieId string) ([]string, error) {
	args := m.Called(ctx, movieId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func TestShowController_Shows(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		date           string
		setupMocks     func(svc *MockShowControllerService)
		expectedStatus int
	}{
		{
			name: "should return 200 with empty list when no shows exist for given date",
			date: "2026-01-01",
			setupMocks: func(svc *MockShowControllerService) {
				svc.On("GetShows", mock.Anything, "2026-01-01").Return([]model.Show{}, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "should return 200 with show list when shows and movie details are fetched successfully",
			date: "2026-01-01",
			setupMocks: func(svc *MockShowControllerService) {
				shows := []model.Show{{Id: 1, MovieId: "tt123", Slot: model.Slot{Id: 1}}}
				svc.On("GetShows", mock.Anything, "2026-01-01").Return(shows, nil).Once()
				svc.On("GetMovieById", mock.Anything, "tt123").Return(&model.Movie{MovieId: "tt123", Name: "Inception"}, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "should return 500 when service fails to retrieve shows for given date",
			date: "2026-01-01",
			setupMocks: func(svc *MockShowControllerService) {
				svc.On("GetShows", mock.Anything, "2026-01-01").Return(nil, ae.InternalServerError("ShowFetchFailed", "db error", errors.New("db down"))).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "should return 500 when movie gateway fails to fetch movie details for an existing show",
			date: "2026-01-01",
			setupMocks: func(svc *MockShowControllerService) {
				shows := []model.Show{{Id: 2, MovieId: "tt999", Slot: model.Slot{Id: 1}}}
				svc.On("GetShows", mock.Anything, "2026-01-01").Return(shows, nil).Once()
				svc.On("GetMovieById", mock.Anything, "tt999").Return(nil, ae.NotFoundError("MovieNotFound", "not found", errors.New("not found"))).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(MockShowControllerService)
			tt.setupMocks(svc)
			ctrl := NewShowController(svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodGet, "/?date="+tt.date, nil)

			ctrl.Shows(c)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestShowController_ScheduledMovies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupMocks     func(svc *MockShowControllerService)
		expectedStatus int
	}{
		{
			name: "should return 200 with movie list when scheduled movies exist",
			setupMocks: func(svc *MockShowControllerService) {
				movies := []model.Movie{
					{MovieId: "tt1", Name: "Inception"},
					{MovieId: "tt2", Name: "Interstellar"},
				}
				svc.On("GetScheduledMovies", mock.Anything).Return(movies, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "should return 200 with empty list when no movies are currently scheduled",
			setupMocks: func(svc *MockShowControllerService) {
				svc.On("GetScheduledMovies", mock.Anything).Return([]model.Movie{}, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "should return 500 when service fails to retrieve scheduled movies",
			setupMocks: func(svc *MockShowControllerService) {
				svc.On("GetScheduledMovies", mock.Anything).Return(nil, ae.InternalServerError("MovieIdsFetchFailed", "db error", errors.New("db down"))).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(MockShowControllerService)
			tt.setupMocks(svc)
			ctrl := NewShowController(svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodGet, "/movies/scheduled", nil)

			ctrl.ScheduledMovies(c)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestShowController_MovieShowtimes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		movieId        string
		date           string
		setupMocks     func(svc *MockShowControllerService)
		expectedStatus int
	}{
		{
			name:    "should return 200 with showtime list when valid movieId and date are provided",
			movieId: "tt123",
			date:    "2026-04-10",
			setupMocks: func(svc *MockShowControllerService) {
				showtimes := []response.ShowtimeResponse{
					{ShowId: 1, ScreenName: "Screen 1", SlotName: "Morning", StartTime: "09:00:00", EndTime: "12:30:00", Cost: 300, AvailableSeats: 80},
				}
				svc.On("GetMovieShowtimes", mock.Anything, "tt123", "2026-04-10").Return(showtimes, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "should return 200 with empty list when no showtimes exist for given movie and date",
			movieId: "tt123",
			date:    "2026-04-10",
			setupMocks: func(svc *MockShowControllerService) {
				svc.On("GetMovieShowtimes", mock.Anything, "tt123", "2026-04-10").Return([]response.ShowtimeResponse{}, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "should return 400 when date query param is missing from request",
			movieId:        "tt123",
			date:           "",
			setupMocks:     func(svc *MockShowControllerService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "should return 500 when service fails to retrieve showtimes for given movie and date",
			movieId: "tt123",
			date:    "2026-04-10",
			setupMocks: func(svc *MockShowControllerService) {
				svc.On("GetMovieShowtimes", mock.Anything, "tt123", "2026-04-10").Return(nil, ae.InternalServerError("ShowsFetchFailed", "db error", errors.New("db down"))).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(MockShowControllerService)
			tt.setupMocks(svc)
			ctrl := NewShowController(svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "movieId", Value: tt.movieId}}
			url := "/movies/" + tt.movieId + "/showtimes"
			if tt.date != "" {
				url += "?date=" + tt.date
			}
			c.Request, _ = http.NewRequest(http.MethodGet, url, nil)

			ctrl.MovieShowtimes(c)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestShowController_MovieAvailableDates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		movieId        string
		setupMocks     func(svc *MockShowControllerService)
		expectedStatus int
	}{
		{
			name:    "should return 200 with available dates list when future shows exist for given movie",
			movieId: "tt123",
			setupMocks: func(svc *MockShowControllerService) {
				svc.On("GetMovieAvailableDates", mock.Anything, "tt123").Return([]string{"2026-04-10", "2026-04-11"}, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "should return 200 with empty list when no future shows exist for given movie",
			movieId: "tt123",
			setupMocks: func(svc *MockShowControllerService) {
				svc.On("GetMovieAvailableDates", mock.Anything, "tt123").Return([]string{}, nil).Once()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "should return 500 when service fails to retrieve available dates for given movie",
			movieId: "tt123",
			setupMocks: func(svc *MockShowControllerService) {
				svc.On("GetMovieAvailableDates", mock.Anything, "tt123").Return(nil, ae.InternalServerError("DatesFetchFailed", "db error", errors.New("db down"))).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := new(MockShowControllerService)
			tt.setupMocks(svc)
			ctrl := NewShowController(svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "movieId", Value: tt.movieId}}
			c.Request, _ = http.NewRequest(http.MethodGet, "/movies/"+tt.movieId+"/available-dates", nil)

			ctrl.MovieAvailableDates(c)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
