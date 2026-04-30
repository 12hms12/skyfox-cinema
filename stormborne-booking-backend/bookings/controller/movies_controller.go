package controller

import (
	"context"
	"net/http"
	"skyfox/bookings/model"
	"skyfox/common/dto/response"
	"skyfox/common/logger"
	ae "skyfox/error"

	"github.com/gin-gonic/gin"
)

type MovieService interface {
	GetAllMovies(ctx context.Context) ([]response.MovieResponse, error)
	GetMovieByID(ctx context.Context, movieID string) (*model.Movie, error)
}

type movieController struct {
	movieService MovieService
}

func NewMovieController(movieService MovieService) *movieController {
	return &movieController{
		movieService: movieService,
	}
}

// Shows godoc
//
//		@Summary		Movie
//		@Description	get shows by date
//		@Tags			Movies
//		@Accept			json
//		@Produce		json
//	 @security	BasicAuth
//	 @param Authorization header string true "Enter basic auth"
//		@Success		200	{object}	[]response.MovieResponse
//		@Failure		400	{object}	ae.AppError
//		@Failure		404	{object}	ae.AppError
//		@Failure		500	{object}	ae.AppError
//		@Router			/movies [get]
func (sh *movieController) Movies(c *gin.Context) {
	shows, responseError := sh.movieService.GetAllMovies(c.Request.Context())
	if responseError != nil {
		err := responseError.(*ae.AppError)
		logger.Error("%s", err.UnWrap().Error())
		c.AbortWithStatusJSON(err.HTTPCode(), err)
	}
	c.IndentedJSON(http.StatusOK, shows)
}

// MovieByID godoc
//
//	@Summary		Movie Details
//	@Description	get movie details by movie id
//	@Tags			Movies
//	@Accept			json
//	@Produce		json
//	@security		BasicAuth
//	@param Authorization header string true "Enter basic auth"
//	@Param movieId path string true "movie id"
//	@Success		200	{object}	model.Movie
//	@Failure		400	{object}	ae.AppError
//	@Failure		404	{object}	ae.AppError
//	@Failure		500	{object}	ae.AppError
//	@Router			/admin/movies/{movieId} [get]
func (sh *movieController) MovieByID(c *gin.Context) {
	movieID := c.Param("movieId")
	movie, responseError := sh.movieService.GetMovieByID(c.Request.Context(), movieID)
	if responseError != nil {
		err := responseError.(*ae.AppError)
		logger.Error("%s", err.UnWrap().Error())
		c.AbortWithStatusJSON(err.HTTPCode(), err)
		return
	}

	c.IndentedJSON(http.StatusOK, movie)
}
