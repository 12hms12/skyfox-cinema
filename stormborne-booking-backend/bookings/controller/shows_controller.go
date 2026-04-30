package controller

import (
	"context"

	"net/http"
	"skyfox/bookings/dto/response"
	"skyfox/bookings/model"
	"skyfox/common/logger"
	ae "skyfox/error"

	"github.com/gin-gonic/gin"
)

type ShowService interface {
	GetShows(context.Context, string) ([]model.Show, error)
	GetMovieById(context.Context, string) (*model.Movie, error)
	GetScheduledMovies(context.Context) ([]model.Movie, error)
	GetMovieShowtimes(context.Context, string, string) ([]response.ShowtimeResponse, error)
	GetMovieAvailableDates(context.Context, string) ([]string, error)
}

type showController struct {
	showService ShowService
}

func NewShowController(showService ShowService) *showController {
	return &showController{
		showService: showService,
	}
}

// Shows godoc
//
//		@Summary		Shows
//		@Description	get shows by date
//		@Tags			Shows
//		@Accept			json
//		@Produce		json
//	 @security	BasicAuth
//	 @param Authorization header string true "Enter basic auth"
//	 @Param  date  query string true "to get shows"
//		@Success		200	{object}	response.ShowResponse
//		@Failure		400	{object}	ae.AppError
//		@Failure		404	{object}	ae.AppError
//		@Failure		500	{object}	ae.AppError
//		@Router			/shows [get]
func (sh *showController) Shows(c *gin.Context) {
	date := c.Request.URL.Query().Get("date")

	shows, responseError := sh.showService.GetShows(c.Request.Context(), date)
	if responseError != nil {
		err := responseError.(*ae.AppError)
		logger.Error("%s", err.UnWrap().Error())
		c.AbortWithStatusJSON(err.HTTPCode(), err)
	}

	var showResponses []response.ShowResponse
	for _, show := range shows {
		show_response, responseError := sh.constructShowResponse(c.Request.Context(), show)
		if responseError != nil {
			err := responseError.(*ae.AppError)
			// logger.Error(err.UnWrap().Error())
			c.AbortWithStatusJSON(err.HTTPCode(), err)
			return
		}
		showResponses = append(showResponses, *show_response)
	}
	c.IndentedJSON(http.StatusOK, showResponses)
}

func (sh *showController) constructShowResponse(ctx context.Context, s model.Show) (*response.ShowResponse, error) {
	movie, err := sh.showService.GetMovieById(ctx, s.MovieId)
	if err != nil {
		return &response.ShowResponse{}, err
	}
	return response.NewShowResponse(*movie, s.Slot, s), nil
}

// ScheduledMovies godoc
//
//	@Summary		Scheduled Movies
//	@Description	get all movies that have upcoming scheduled shows
//	@Tags			Shows
//	@Produce		json
//	@Success		200	{array}		model.Movie
//	@Failure		500	{object}	ae.AppError
//	@Router			/movies/scheduled [get]
func (sh *showController) ScheduledMovies(c *gin.Context) {
	movies, responseError := sh.showService.GetScheduledMovies(c.Request.Context())
	if responseError != nil {
		err := responseError.(*ae.AppError)
		logger.Error("%s", err.UnWrap().Error())
		c.AbortWithStatusJSON(err.HTTPCode(), err)
		return
	}
	c.IndentedJSON(http.StatusOK, movies)
}

// MovieShowtimes godoc
//
//	@Summary		Movie Showtimes
//	@Description	get all showtimes for a movie on a given date
//	@Tags			Shows
//	@Produce		json
//	@Param			movieId	path		string	true	"movie id"
//	@Param			date	query		string	true	"date (YYYY-MM-DD)"
//	@Success		200	{array}		response.ShowtimeResponse
//	@Failure		400	{object}	ae.AppError
//	@Failure		500	{object}	ae.AppError
//	@Router			/movies/{movieId}/showtimes [get]
func (sh *showController) MovieShowtimes(c *gin.Context) {
	movieId := c.Param("movieId")
	date := c.Query("date")

	if date == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "date query parameter is required"})
		return
	}

	showtimes, responseError := sh.showService.GetMovieShowtimes(c.Request.Context(), movieId, date)
	if responseError != nil {
		err := responseError.(*ae.AppError)
		logger.Error("%s", err.UnWrap().Error())
		c.AbortWithStatusJSON(err.HTTPCode(), err)
		return
	}
	c.IndentedJSON(http.StatusOK, showtimes)
}

// MovieAvailableDates godoc
//
//	@Summary		Movie Available Dates
//	@Description	get all dates that have shows for a given movie
//	@Tags			Shows
//	@Produce		json
//	@Param			movieId	path		string	true	"movie id"
//	@Success		200	{array}		string
//	@Failure		500	{object}	ae.AppError
//	@Router			/movies/{movieId}/available-dates [get]
func (sh *showController) MovieAvailableDates(c *gin.Context) {
	movieId := c.Param("movieId")

	dates, responseError := sh.showService.GetMovieAvailableDates(c.Request.Context(), movieId)
	if responseError != nil {
		err := responseError.(*ae.AppError)
		logger.Error("%s", err.UnWrap().Error())
		c.AbortWithStatusJSON(err.HTTPCode(), err)
		return
	}
	c.IndentedJSON(http.StatusOK, dates)
}
