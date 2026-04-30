package integrationtest

import (
	"context"
	"fmt"
	"net/http"
	"skyfox/bookings/constants"
	"skyfox/bookings/controller"
	"skyfox/bookings/model"
	"skyfox/bookings/repository"
	repoMocks "skyfox/bookings/repository/mocks"
	"skyfox/bookings/service"
	"skyfox/common/middleware/security"
	db "skyfox/integration_test/db"
	"skyfox/movieservice/movie_gateway/mocks"
	testdata "skyfox/test_data"
	"testing"
	"time"

	"github.com/appleboy/gofight/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var revenuePath = constants.RevenueEndPoint

func Test_WhenGetRevenue_ItShouldReturnRevenue(t *testing.T) {
    tearDown := revenueControllerTestSetup()
    defer tearDown(t)

	testUser := testdata.DummyUsers[0]
    testUserToken := testdata.GetTestJwtToken(testUser)

    db := db.GetDB()
    ctx := context.Background()

    mockMovieGateway := mocks.NewMockMovieGateWay(t)
    mockSeatRepo := repoMocks.NewMockSeatRepository(t)
    
    showID := 1
    seatIDs := []int{1, 2}
    totalPrice := 600.0

    bookingRepo := repository.NewCustomerBookingRepository(db)
    
    booking, err := bookingRepo.CreateBookingWithSeatLock(ctx, int64(testUser.ID), showID, seatIDs, totalPrice, time.Now().Add(time.Minute), "transaction_123")
    assert.NoError(t, err)
    err = bookingRepo.ConfirmBooking(ctx, booking.ID)
    assert.NoError(t, err)

    bookingRepoGeneral := repository.NewBookingRepository(db)
    showRepo := repository.NewShowRepository(db, mockMovieGateway, mockSeatRepo)
    revenueService := service.NewRevenueService(bookingRepoGeneral, showRepo)
    handler := controller.NewRevenueController(revenueService)

    engine, request := getEngine()
    engine.GET(revenuePath, security.JWTAuth(testdata.DummyJwtManager, security.USER), handler.GetRevenue)

    mockMovieGateway.On("MovieById", mock.Anything, "tt6857189").Return(&model.Movie{MovieId: "tt6857189", Name: "Movie One"}, nil)
	mockMovieGateway.On("MovieById", mock.Anything, "tt6856489").Return(&model.Movie{MovieId: "tt6856489", Name: "Movie two"}, nil)
	mockMovieGateway.On("MovieById", mock.Anything, "tt6856999").Return(&model.Movie{MovieId: "tt6856999", Name: "Movie three"}, nil)

    expectedResponse := `
    {
        "grossRevenue": 600,
        "shows": [
            {
                "id": 1,
                "date": "2022-10-13",
                "showTime": "18:00:00",
                "movie": {
                    "id": "tt6857189",
                    "name": "Movie One",
                    "duration": "",
                    "plot": "",
                    "poster": "",
                    "genre": "",
                    "imdbRating": "",
                    "imdbVotes":"",
                    "rated":""
                },
                "revenue": 600,
                "ticketSold": 2
            }
        ]
    }`

    request.GET(revenuePath).
        SetQuery(gofight.H{"startDate": "2022-10-13"}).
        SetQuery(gofight.H{"endDate": "2022-10-13"}).
        SetHeader(gofight.H{"Authorization": fmt.Sprintf("Bearer %s", testUserToken)}).
        Run(engine, func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
            assert.Equal(t, http.StatusOK, r.Code)
            assert.JSONEq(t, expectedResponse, r.Body.String())
        })
}
func Test_WhenGetRevenue_ForNoBooking_ItShouldReturnZero(t *testing.T) {

	tearDown := revenueControllerTestSetup()
	defer tearDown(t)

	testUser := testdata.DummyUsers[0]
    testUserToken := testdata.GetTestJwtToken(testUser)

	db := db.GetDB()
	ctx := context.Background()

	mockMovieGateway := mocks.NewMockMovieGateWay(t)
	mockSeatRepo := repoMocks.NewMockSeatRepository(t)


	handler := controller.NewRevenueController(service.NewRevenueService(repository.NewBookingRepository(db), repository.NewShowRepository(db, mockMovieGateway, mockSeatRepo)))

	engine, request := getEngine()
	engine.GET(revenuePath, security.JWTAuth(testdata.DummyJwtManager, security.USER), handler.GetRevenue)

	mockMovieGateway.On("MovieById", ctx, mock.AnythingOfType("string")).Return(&model.Movie{}, nil)

	expectedResponse := `
	{
		"grossRevenue": 0,
		"shows": []
	}`

	request.GET(revenuePath).
		SetDebug(true).SetHeader(gofight.H{"Authorization": fmt.Sprintf("Bearer %s", testUserToken)}).
		Run(engine, func(r gofight.HTTPResponse, rq gofight.HTTPRequest) {
			assert.Equal(t, http.StatusOK, r.Code)
			assert.JSONEq(t, expectedResponse, r.Body.String())
		})
}

func revenueControllerTestSetup() func(*testing.T) {
    db := db.GetDB()
    gormDB := db.GormDB()

    tables := []string{"booked_seat", "show_seat_status", "show_pricing", "booking", "show", "slot", "screen", "customer"}
    for _, table := range tables {
        gormDB.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
    }

    screen := model.Screen{ScreenName: "test"}
    gormDB.Create(&screen)

    gormDB.Create(testdata.DummyShows)

    var seats []model.Seat
    for row := 1; row <= 10; row++ {
        for col := 1; col <= 10; col++ {
            seat := model.Seat{
                ScreenID:     screen.ID,
                RowNumber:    row,
                ColumnNumber: col,
                SeatType:     "REGULAR",
            }
            gormDB.Create(&seat)
            seats = append(seats, seat)
        }
    }

    for _, show := range testdata.DummyShows {
        gormDB.Create(&model.ShowPricing{
            ShowID:   show.Id,
            SeatType: "REGULAR",
            Price:    show.Cost,
        })

        for _, seat := range seats {
            status := "AVAILABLE"
            gormDB.Create(&model.ShowSeatStatus{
                ShowID: show.Id,
                SeatID: seat.ID,
                Status: status,
            })
        }
    }

    gormDB.Create(&testdata.DummyUsers)

    return func(t *testing.T) {
        for _, table := range tables {
            gormDB.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
        }
    }
}