package controller

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"skyfox/bookings/model"
	commonResponse "skyfox/common/dto/response"
	ae "skyfox/error"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMoviesControllerService struct {
	mock.Mock
}

func (m *MockMoviesControllerService) GetAllMovies(ctx context.Context) ([]commonResponse.MovieResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]commonResponse.MovieResponse), args.Error(1)
}

func (m *MockMoviesControllerService) GetMovieByID(ctx context.Context, movieID string) (*model.Movie, error) {
	args := m.Called(ctx, movieID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Movie), args.Error(1)
}

func TestMoviesController_Movies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns movie list", func(t *testing.T) {
		svc := new(MockMoviesControllerService)
		ctrl := NewMovieController(svc)
		svc.On("GetAllMovies", mock.Anything).Return([]commonResponse.MovieResponse{{Id: "tt1", Title: "Inception"}}, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/movies", nil)

		ctrl.Movies(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("returns service error", func(t *testing.T) {
		svc := new(MockMoviesControllerService)
		ctrl := NewMovieController(svc)
		svc.On("GetAllMovies", mock.Anything).Return(nil, ae.InternalServerError("InternalServerError", "failed", errors.New("failed"))).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/movies", nil)

		ctrl.Movies(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestMoviesController_MovieByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns movie details", func(t *testing.T) {
		svc := new(MockMoviesControllerService)
		ctrl := NewMovieController(svc)
		svc.On("GetMovieByID", mock.Anything, "tt123").Return(&model.Movie{MovieId: "tt123", Name: "Inception", Duration: "2h28m0s"}, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "movieId", Value: "tt123"}}
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/movies/tt123", nil)

		ctrl.MovieByID(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("returns service error", func(t *testing.T) {
		svc := new(MockMoviesControllerService)
		ctrl := NewMovieController(svc)
		svc.On("GetMovieByID", mock.Anything, "tt404").Return(nil, ae.NotFoundError("MovieNotFound", "movie not found", errors.New("not found"))).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "movieId", Value: "tt404"}}
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/movies/tt404", nil)

		ctrl.MovieByID(c)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
