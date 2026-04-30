package movieservice

import (
	"context"
	"skyfox/bookings/constants"
	"skyfox/bookings/model"
	"skyfox/config"
	"skyfox/movieservice/movie_gateway/mocks"
	"testing"

	"fmt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/h2non/gock.v1"
)

type errorReader struct {
	data []byte
	read bool
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	if !e.read {
		n = copy(p, e.data)
		e.read = true
		return n, nil
	}
	return 0, fmt.Errorf("forced read error")
}
func Test_ReturnsMovie_When_MovieServiceIsInvoked(t *testing.T) {
	posterURL := "https://m.media-amazon.com/images/M/MV5BMjI0MDMzNTQ0M15BMl5BanBnXkFtZTgwMTM5NzM3NDM@._V1_SX300.jpg"

	want := &model.Movie{
		MovieId:    "movie_id",
		Name:       "movie_name",
		Duration:   "1h30m0s",
		Plot:       "movie plot in short",
		Poster:     posterURL,
		Genre:      "Drama, Horror, Sci-Fi",
		ImdbRating: "7.5",
		Rated: "PG-13",
		ImdbVotes: "379,472",
	}

	movieGatewayRepo := mocks.MockMovieGateWay{}
	movieGatewayRepo.On("MovieById", mock.AnythingOfType("string")).Return(want, nil).Once()
	defer gock.Off()
	body := `{"Title":"movie_name","Year":"2018","Rated":"PG-13","Released":"06 Apr 2018","Runtime":"90 min","Genre":"Drama, Horror, Sci-Fi","Director":"John Krasinski","Writer":"Bryan Woods (screenplay by), Scott Beck (screenplay by), John Krasinski (screenplay by), Bryan Woods (story by), Scott Beck (story by)","Actors":"Emily Blunt, John Krasinski, Millicent Simmonds, Noah Jupe","Plot":"movie plot in short","Language":"English, American Sign Language","Country":"USA","Awards":"Nominated for 1 Oscar. Another 34 wins & 108 nominations.","Poster":"https://m.media-amazon.com/images/M/MV5BMjI0MDMzNTQ0M15BMl5BanBnXkFtZTgwMTM5NzM3NDM@._V1_SX300.jpg","Ratings":[{"Source":"Internet Movie Database","Value":"7.5/10"},{"Source":"Rotten Tomatoes","Value":"95%"},{"Source":"Metacritic","Value":"82/100"}],"Metascore":"82","imdbRating":"7.5","imdbVotes":"379,472","imdbID":"movie_id","Type":"movie","DVD":"N/A","BoxOffice":"N/A","Production":"N/A","Website":"N/A","Response":"True"}`
	gock.New("http://localhost:4567").Get("/movies/*").Persist().Reply(200).BodyString(body)

	movieGateway := NewMovieGateway(config.MovieGatewayConfig{MovieServiceHost: "http://localhost:4567/"})

	got, err := movieGateway.MovieById(context.Background(), "")

	assert.Nil(t, err)
	assert.Equal(t, want, got)
}

func Test_ReturnsMoviePoster_When_MovieIDIsProvided(t *testing.T) {
	host := "http://localhost:4567"
	gateway := NewMovieGateway(config.MovieGatewayConfig{MovieServiceHost: host + "/"})
	ctx := context.Background()

	t.Run("Should return Success WHen Full Movie Mapping", func(t *testing.T) {
		defer gock.Off()

		posterBody := "https://m.media-amazon.com"
		body := `{"imdbid":"movie_id","title":"movie_name","runtime":"90 min","plot":"plot","poster":"` + posterBody + `"}`

		gock.New(host).Get("/movies/movie_id").Reply(200).BodyString(body)

		got, err := gateway.MovieById(ctx, "movie_id")

		assert.Nil(t, err)
		assert.Equal(t, posterBody, got.Poster)
	})

	t.Run("Should return Stock Image When Poster is N/A", func(t *testing.T) {
		defer gock.Off()

		body := `{"imdbid":"123","title":"Us","runtime":"116 min","poster":"N/A"}`

		gock.New(host).Get("/movies/123").Reply(200).BodyString(body)

		got, err := gateway.MovieById(ctx, "123")

		assert.Nil(t, err)
		assert.Equal(t, constants.StockImage, got.Poster)
	})

	t.Run("Should return Error Invalid Runtime Format When Invalid Date format is Provided ", func(t *testing.T) {
		defer gock.Off()

		body := `{"imdbid":"123","runtime":"abc min"}`

		gock.New(host).Get("/movies/123").Reply(200).BodyString(body)

		_, err := gateway.MovieById(ctx, "123")

		assert.NotNil(t, err)
	})

	t.Run("Should return Error Invalid JSON Body when Invalid JSON Body", func(t *testing.T) {
		defer gock.Off()
		gock.New(host).Get("/movies/123").Reply(200).BodyString(`{ "invalid": }`)

		_, err := gateway.MovieById(ctx, "123")

		assert.NotNil(t, err)
	})

	t.Run("SHould return Error HTTP Request Failure when HTTP request fails", func(t *testing.T) {
		defer gock.Off()
		gock.New(host).Get("/movies/123").ReplyError(fmt.Errorf("network down"))

		_, err := gateway.MovieById(ctx, "123")

		assert.NotNil(t, err)
	})

	t.Run("Should Error Invalid Runtime when trigger failure", func(t *testing.T) {
		defer gock.Off()
		gock.New("http://localhost:4567").
			Get("/movies/123").
			Reply(200).
			BodyString(`{"imdbid":"123","runtime":"invalid"}`)

		gateway := NewMovieGateway(config.MovieGatewayConfig{MovieServiceHost: "http://localhost:4567/"})
		_, err := gateway.MovieById(context.Background(), "123")

		assert.NotNil(t, err)
	})

	t.Run("Should return Invalid URL when Invalid URL is hit", func(t *testing.T) {
		badGateway := NewMovieGateway(config.MovieGatewayConfig{
			MovieServiceHost: " http://bad-url.com",
		})
		_, err := badGateway.MovieById(context.Background(), "123")
		assert.NotNil(t, err)
	})

	t.Run("Should return could not retrieve the movie detail when Body Read Failure ", func(t *testing.T) {
		defer gock.Off()
		host := "http://localhost:4567"

		gock.New(host).
			Get("/movies/123").
			Reply(200).
			Body(&errorReader{data: []byte("{")})

		gateway := NewMovieGateway(config.MovieGatewayConfig{MovieServiceHost: host + "/"})
		_, err := gateway.MovieById(context.Background(), "123")

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "could not retrieve the movie detail")
	})

	t.Run("Should return Invalid runtime format when Empty Runtime", func(t *testing.T) {
		defer gock.Off()
		gock.New("http://localhost:4567").
			Get("/movies/123").
			Reply(200).
			BodyString(`{"imdbid":"123","runtime":""}`)

		gateway := NewMovieGateway(config.MovieGatewayConfig{MovieServiceHost: "http://localhost:4567/"})
		_, err := gateway.MovieById(context.Background(), "123")

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "Invalid runtime format")
	})

	t.Run("Should return could not parse movie service url when URL Parse Failure", func(t *testing.T) {
		badGateway := NewMovieGateway(config.MovieGatewayConfig{
			MovieServiceHost: "http:// bad url.com",
		})
		_, err := badGateway.MovieById(context.Background(), "123")

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "could not parse movie service url")
	})
}

func Test_GetAllMovies(t *testing.T) {
	host := "http://localhost:4567"
	gateway := NewMovieGateway(config.MovieGatewayConfig{MovieServiceHost: host + "/"})
	ctx := context.Background()

	t.Run("Should return list of movies when Movie Service returns valid JSON array", func(t *testing.T) {
		defer gock.Off()
		body := `[{"imdbid":"1","title":"Movie A"},{"imdbid":"2","title":"Movie B"}]`
		gock.New(host).Get("/movies").Reply(200).BodyString(body)

		got, err := gateway.GetAllMovies(ctx)

		assert.Nil(t, err)
		assert.Equal(t, 2, len(got))
	})

	t.Run("Should return could not retrieve the movie detail when network fails", func(t *testing.T) {
		defer gock.Off()
		gock.New(host).Get("/movies").ReplyError(fmt.Errorf("timeout"))

		_, err := gateway.GetAllMovies(ctx)

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "could not retrieve the movie detail")
	})

	t.Run("Should return failed to parse the movie details when JSON is not an array", func(t *testing.T) {
		defer gock.Off()

		gock.New(host).Get("/movies").Reply(200).BodyString(`{"error": "not an array"}`)

		_, err := gateway.GetAllMovies(ctx)

		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "failed to parse the movie details")
	})
}
