package movieservice

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"skyfox/bookings/model"
	"skyfox/common/dto/response"
	"skyfox/config"

	ae "skyfox/error"

	"github.com/bborbe/http/requestbuilder"
)

type MovieGateWay interface {
	MovieById(ctx context.Context, id string) (*model.Movie, error)
	GetAllMovies(ctx context.Context) ([]response.MovieResponse, error)
}

type movieGateway struct {
	config config.MovieGatewayConfig
	client *http.Client
}

func NewMovieGateway(cfg config.MovieGatewayConfig) MovieGateWay {
	return &movieGateway{
		config: config.MovieGatewayConfig{
			MovieServiceHost: cfg.MovieServiceHost,
		},
		client: &http.Client{},
	}
}

func (d *movieGateway) MovieById(ctx context.Context, id string) (*model.Movie, error) {

	var err error

	request, err := requestbuilder.NewHTTPRequestBuilder(d.config.MovieServiceHost + "movies/" + id).Build()
	if err != nil {
		return &model.Movie{}, ae.InternalServerError("InternalServerError", "could not parse movie service url", err)
	}

	httpResponse, err := d.client.Do(request)
	if err != nil {
		return &model.Movie{}, ae.InternalServerError("InternalServerError", "could not retrieve the movie detail", err)
	}
	defer httpResponse.Body.Close()
	responseBody, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		return &model.Movie{}, ae.InternalServerError("InternalServerError", "failed to read the movie response body", err)
	}

	var movieResponse MovieServiceResponse
	err = json.Unmarshal(responseBody, &movieResponse)

	if err != nil {
		return &model.Movie{}, ae.InternalServerError("InternalServerError", "failed to parse the movie details", err)
	}

	movie, err := movieResponse.ToMovie()
	if err != nil {
		return &model.Movie{}, err
	}
	return movie, nil
}

func (d *movieGateway) GetAllMovies(ctx context.Context) ([]response.MovieResponse, error) {

	var err error

	request, err := requestbuilder.NewHTTPRequestBuilder(d.config.MovieServiceHost + "movies").Build()
	if err != nil {
		return []response.MovieResponse{}, ae.InternalServerError("InternalServerError", "could not parse movie service url", err)
	}

	httpResponse, err := d.client.Do(request)
	if err != nil {
		return []response.MovieResponse{}, ae.InternalServerError("InternalServerError", "could not retrieve the movie detail", err)
	}
	defer httpResponse.Body.Close()
	responseBody, _ := ioutil.ReadAll(httpResponse.Body)

	var getMoviesResponse []response.MovieResponse
	err = json.Unmarshal(responseBody, &getMoviesResponse)

	if err != nil {
		return []response.MovieResponse{}, ae.InternalServerError("InternalServerError", "failed to parse the movie details", err)
	}
	return getMoviesResponse, nil
}
