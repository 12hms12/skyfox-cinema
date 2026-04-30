package service

import (
	"context"
	"skyfox/bookings/model"
	"skyfox/common/dto/response"
	movieservice "skyfox/movieservice/movie_gateway"
)

type movieService struct {
	movieGateway movieservice.MovieGateWay
}

func NewMovieService(gateway movieservice.MovieGateWay) *movieService {
	return &movieService{
		movieGateway: gateway,
	}
}

func (s *movieService) GetAllMovies(ctx context.Context) ([]response.MovieResponse, error) {
	movies, err := s.movieGateway.GetAllMovies(ctx)
	if err != nil {
		return []response.MovieResponse{}, err
	}
	return movies, nil
}

func (s *movieService) GetMovieByID(ctx context.Context, movieID string) (*model.Movie, error) {
	movie, err := s.movieGateway.MovieById(ctx, movieID)
	if err != nil {
		return &model.Movie{}, err
	}
	return movie, nil
}


