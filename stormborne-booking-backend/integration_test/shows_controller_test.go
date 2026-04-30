package integrationtest

import (
	"net/http"
	"skyfox/bookings/constants"
	"skyfox/bookings/controller"
	"skyfox/bookings/model"
	"skyfox/bookings/repository"
	repoMocks "skyfox/bookings/repository/mocks"
	"skyfox/bookings/service"
	"skyfox/config"
	db "skyfox/integration_test/db"
	movieservice "skyfox/movieservice/movie_gateway"
	"skyfox/movieservice/movie_gateway/mocks"
	testdata "skyfox/test_data"
	"testing"

	"github.com/appleboy/gofight/v2"
	"gopkg.in/h2non/gock.v1"
	"gotest.tools/assert"
)

var showsPath = constants.ShowEndPoint

// TODO - mark it as test with Test_
func Test_WhenGetShows_ItShouldReturnAllShowsForTheDate(t *testing.T) {
	db := db.GetDB()
	gormDB := db.GormDB()


	screen := model.Screen{ScreenName: "test"}
    gormDB.Create(&screen)

    gormDB.Create(testdata.DummyShows)

	mockMovieService := mocks.NewMockMovieGateWay(t)
    mockSeatRepo := repoMocks.NewMockSeatRepository(t)

	defer gock.Off()
	body := `{"Title":"A Quiet Place","Year":"2018","Rated":"PG-13","Released":"06 Apr 2018","Runtime":"90 min","Genre":"Drama, Horror, Sci-Fi","Director":"John Krasinski","Writer":"Bryan Woods (screenplay by), Scott Beck (screenplay by), John Krasinski (screenplay by), Bryan Woods (story by), Scott Beck (story by)","Actors":"Emily Blunt, John Krasinski, Millicent Simmonds, Noah Jupe","Plot":"In a post-apocalyptic world, a family is forced to live in silence while hiding from monsters with ultra-sensitive hearing.","Language":"English, American Sign Language","Country":"USA","Awards":"Nominated for 1 Oscar. Another 34 wins & 108 nominations.","Poster":"https://m.media-amazon.com/images/M/MV5BMjI0MDMzNTQ0M15BMl5BanBnXkFtZTgwMTM5NzM3NDM@._V1_SX300.jpg","Ratings":[{"Source":"Internet Movie Database","Value":"7.5/10"},{"Source":"Rotten Tomatoes","Value":"95%"},{"Source":"Metacritic","Value":"82/100"}],"Metascore":"82","imdbRating":"7.5","imdbVotes":"379,472","imdbID":"tt6644200","Type":"movie","DVD":"N/A","BoxOffice":"N/A","Production":"N/A","Website":"N/A","Response":"True"}`
	gock.New("http://localhost:4567").Get("/movies/*").Persist().Reply(200).BodyString(body)

	// Review - gock to mock the movie service request

	handler := controller.NewShowController(service.NewShowService(repository.NewShowRepository(db,mockMovieService,mockSeatRepo), movieservice.NewMovieGateway(config.MovieGatewayConfig{MovieServiceHost: "http://localhost:4567/"})))

	engine, request := getEngine()
	engine.GET(showsPath, handler.Shows)

	request.GET(showsPath).SetDebug(true).
		SetQuery(gofight.H{"date": "2022-10-13"}).
		Run(engine, func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			assert.Equal(t, http.StatusOK, r.Code)

			assert.Equal(t, res, r.Body.String())
		})
}

var res = `[
    {
        "movie": {
            "id": "tt6644200",
            "name": "A Quiet Place",
            "duration": "1h30m0s",
            "plot": "In a post-apocalyptic world, a family is forced to live in silence while hiding from monsters with ultra-sensitive hearing.",
            "poster": "https://m.media-amazon.com/images/M/MV5BMjI0MDMzNTQ0M15BMl5BanBnXkFtZTgwMTM5NzM3NDM@._V1_SX300.jpg",
            "genre": "Drama, Horror, Sci-Fi",
            "imdbRating": "7.5",
            "rated": "PG-13",
            "imdbVotes": "379,472"
        },
        "slot": {
            "id": 3,
            "name": "slot3",
            "startTime": "18:00:00",
            "endTime": "21:30:00"
        },
        "id": 1,
        "date": "2022-10-13",
        "cost": 300
    },
    {
        "movie": {
            "id": "tt6644200",
            "name": "A Quiet Place",
            "duration": "1h30m0s",
            "plot": "In a post-apocalyptic world, a family is forced to live in silence while hiding from monsters with ultra-sensitive hearing.",
            "poster": "https://m.media-amazon.com/images/M/MV5BMjI0MDMzNTQ0M15BMl5BanBnXkFtZTgwMTM5NzM3NDM@._V1_SX300.jpg",
            "genre": "Drama, Horror, Sci-Fi",
            "imdbRating": "7.5",
            "rated": "PG-13",
            "imdbVotes": "379,472"
        },
        "slot": {
            "id": 4,
            "name": "slot4",
            "startTime": "22:30:00",
            "endTime": "02:00:00"
        },
        "id": 2,
        "date": "2022-10-13",
        "cost": 350
    },
    {
        "movie": {
            "id": "tt6644200",
            "name": "A Quiet Place",
            "duration": "1h30m0s",
            "plot": "In a post-apocalyptic world, a family is forced to live in silence while hiding from monsters with ultra-sensitive hearing.",
            "poster": "https://m.media-amazon.com/images/M/MV5BMjI0MDMzNTQ0M15BMl5BanBnXkFtZTgwMTM5NzM3NDM@._V1_SX300.jpg",
            "genre": "Drama, Horror, Sci-Fi",
            "imdbRating": "7.5",
            "rated": "PG-13",
            "imdbVotes": "379,472"
        },
        "slot": {
            "id": 1,
            "name": "slot1",
            "startTime": "09:00:00",
            "endTime": "12:30:00"
        },
        "id": 3,
        "date": "2022-10-13",
        "cost": 350
    }
]`
