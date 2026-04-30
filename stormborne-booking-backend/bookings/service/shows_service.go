package service

import (
	"context"
	"skyfox/bookings/dto/response"
	"skyfox/bookings/model"
	"skyfox/bookings/repository"
	"skyfox/common/logger"
	movieservice "skyfox/movieservice/movie_gateway"
	"sort"
	"sync"
)

type ShowRepository interface {
	GetAllShowsOn(ctx context.Context, date string) ([]model.Show, error)
	FindById(ctx context.Context, id int) (model.Show, error)
	GetAllShowsBy(ctx context.Context, revenueQuery *repository.RevenueQuery) ([]model.Show, error)
	GetScheduledMovieIds(ctx context.Context) ([]string, error)
	GetShowsByMovieAndDate(ctx context.Context, movieId, date string) ([]model.Show, error)
	GetAvailableDatesForMovie(ctx context.Context, movieId string) ([]string, error)
	CountAvailableSeats(ctx context.Context, showId int) (int64, error)
	CreateScreen(ctx context.Context, screen *model.Screen) error 
}

type showService struct {
	showRepo     ShowRepository
	movieGateway movieservice.MovieGateWay
}

func NewShowService(showRepository ShowRepository, gateway movieservice.MovieGateWay) *showService {
	return &showService{
		showRepo:     showRepository,
		movieGateway: gateway,
	}
}

func (s *showService) GetShows(ctx context.Context, date string) ([]model.Show, error) {
	shows, err := s.showRepo.GetAllShowsOn(ctx, date)
	if err != nil {
		return nil, err
	}
	for i := range shows {
		movie, err := s.movieGateway.MovieById(ctx, shows[i].MovieId)
		if err == nil {
			shows[i].Movie = *movie
		}
	}

	return shows, nil
}

func (s *showService) GetMovieById(ctx context.Context, movieId string) (*model.Movie, error) {
	movie, err := s.movieGateway.MovieById(ctx, movieId)
	if err != nil {
		return &model.Movie{}, err
	}
	return movie, nil
}

func (s *showService) GetScheduledMovies(ctx context.Context) ([]model.Movie, error) {
	movieIds, err := s.showRepo.GetScheduledMovieIds(ctx)
	if err != nil {
		return nil, err
	}

	movies := make([]model.Movie, 0, len(movieIds))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, id := range movieIds {
		wg.Add(1)
		go func(movieId string) {
			defer wg.Done()
			movie, err := s.movieGateway.MovieById(ctx, movieId)
			if err != nil {
				logger.Error("failed to fetch movie details for id: %s", movieId)
				return
			}
			mu.Lock()
			movies = append(movies, *movie)
			mu.Unlock()
		}(id)
	}
	wg.Wait()

	sort.Slice(movies, func(i, j int) bool {
		return movies[i].MovieId < movies[j].MovieId
	})

	return movies, nil
}

func (s *showService) GetMovieShowtimes(ctx context.Context, movieId, date string) ([]response.ShowtimeResponse, error) {
	shows, err := s.showRepo.GetShowsByMovieAndDate(ctx, movieId, date)
	if err != nil {
		return nil, err
	}

	showtimes := make([]response.ShowtimeResponse, 0, len(shows))
	for _, show := range shows {
		availableSeats, err := s.showRepo.CountAvailableSeats(ctx, show.Id)
		if err != nil {
			availableSeats = 0
		}

		showtimes = append(showtimes, response.ShowtimeResponse{
			ShowId:         show.Id,
			ScreenId:       show.ScreenID,
			ScreenName:     show.Screen.ScreenName,
			SlotId:         show.SlotId,
			SlotName:       show.Slot.Name,
			StartTime:      show.Slot.StartTime,
			EndTime:        show.Slot.EndTime,
			Cost:           show.Cost,
			AvailableSeats: availableSeats,
		})
	}

	return showtimes, nil
}

func (s *showService) GetMovieAvailableDates(ctx context.Context, movieId string) ([]string, error) {
	return s.showRepo.GetAvailableDatesForMovie(ctx, movieId)
}