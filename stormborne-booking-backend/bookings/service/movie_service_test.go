package service

import (
	"context"
	"errors"
	"testing"

	"skyfox/bookings/model"
	commonResponse "skyfox/common/dto/response"
	movieGatewayMock "skyfox/movieservice/movie_gateway/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMovieService_GetAllMovies(t *testing.T) {
	gateway := movieGatewayMock.NewMockMovieGateWay(t)
	svc := NewMovieService(gateway)

	expected := []commonResponse.MovieResponse{{Id: "tt1", Title: "Inception"}}
	gateway.On("GetAllMovies", mock.Anything).Return(expected, nil).Once()

	got, err := svc.GetAllMovies(context.Background())

	assert.Nil(t, err)
	assert.Equal(t, expected, got)
}

func TestMovieService_GetMovieByID(t *testing.T) {
	t.Run("returns movie when gateway succeeds", func(t *testing.T) {
		gateway := movieGatewayMock.NewMockMovieGateWay(t)
		svc := NewMovieService(gateway)

		expected := &model.Movie{MovieId: "tt123", Name: "Inception", Duration: "2h28m0s"}
		gateway.On("MovieById", mock.Anything, "tt123").Return(expected, nil).Once()

		got, err := svc.GetMovieByID(context.Background(), "tt123")

		assert.Nil(t, err)
		assert.Equal(t, expected, got)
	})

	t.Run("returns empty movie and error when gateway fails", func(t *testing.T) {
		gateway := movieGatewayMock.NewMockMovieGateWay(t)
		svc := NewMovieService(gateway)

		gateway.On("MovieById", mock.Anything, "tt404").Return(&model.Movie{}, errors.New("not found")).Once()

		got, err := svc.GetMovieByID(context.Background(), "tt404")

		assert.NotNil(t, err)
		assert.Equal(t, &model.Movie{}, got)
	})
}
