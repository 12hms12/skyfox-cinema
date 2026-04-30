package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"skyfox/bookings/dto/request"
	"skyfox/bookings/model"
	ae "skyfox/error"
	movieGatewayMock "skyfox/movieservice/movie_gateway/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAdminScheduleRepository struct {
	mock.Mock
}

func (m *MockAdminScheduleRepository) GetShowsInRange(ctx context.Context, startDate time.Time, endDate time.Time) ([]model.Show, error) {
	args := m.Called(ctx, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Show), args.Error(1)
}

func (m *MockAdminScheduleRepository) GetAllSlots(ctx context.Context) ([]model.Slot, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Slot), args.Error(1)
}

func (m *MockAdminScheduleRepository) GetAllScreens(ctx context.Context) ([]model.Screen, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Screen), args.Error(1)
}

func (m *MockAdminScheduleRepository) GetSlotByID(ctx context.Context, slotID int) (*model.Slot, error) {
	args := m.Called(ctx, slotID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Slot), args.Error(1)
}

func (m *MockAdminScheduleRepository) IsSlotOccupied(ctx context.Context, screenID int, date time.Time, slotID int) (bool, error) {
	args := m.Called(ctx, screenID, date, slotID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAdminScheduleRepository) CreateShow(ctx context.Context, show *model.Show) error {
	args := m.Called(ctx, show)
	return args.Error(0)
}

func (m *MockAdminScheduleRepository) CreateScreen(ctx context.Context, screen *model.Screen) error {
	args := m.Called(ctx, screen)
	return args.Error(0)
}

func (m *MockAdminScheduleRepository) DeleteShowByID(ctx context.Context, showID int) error {
	args := m.Called(ctx, showID)
	return args.Error(0)
}

func TestAdminScheduleService_GetWeeklySchedule(t *testing.T) {
	t.Run("returns bad request when startDate is invalid", func(t *testing.T) {
		repo := new(MockAdminScheduleRepository)
		gateway := movieGatewayMock.NewMockMovieGateWay(t)
		svc := NewAdminScheduleService(repo, gateway)

		resp, err := svc.GetWeeklySchedule(context.Background(), "03-17-2026")

		assert.Nil(t, resp)
		assert.NotNil(t, err)
		appErr := err.(*ae.AppError)
		assert.Equal(t, "InvalidStartDate", appErr.Code)
	})

	t.Run("returns weekly schedule sorted by date/screen/slot", func(t *testing.T) {
		repo := new(MockAdminScheduleRepository)
		gateway := movieGatewayMock.NewMockMovieGateWay(t)
		svc := NewAdminScheduleService(repo, gateway)

		startDate := time.Date(2026, time.March, 16, 0, 0, 0, 0, time.UTC)
		endDate := startDate.AddDate(0, 0, 7)

		shows := []model.Show{
			{MovieId: "tt3", ScreenID: 2, SlotId: 2, Date: time.Date(2026, time.March, 17, 0, 0, 0, 0, time.UTC), Cost: 320},
			{MovieId: "tt1", ScreenID: 1, SlotId: 3, Date: time.Date(2026, time.March, 16, 0, 0, 0, 0, time.UTC), Cost: 280},
			{MovieId: "tt2", ScreenID: 1, SlotId: 1, Date: time.Date(2026, time.March, 16, 0, 0, 0, 0, time.UTC), Cost: 300},
		}

		repo.On("GetShowsInRange", mock.Anything, startDate, endDate).Return(shows, nil).Once()
		gateway.On("MovieById", mock.Anything, "tt1").Return(&model.Movie{Name: "Movie 1"}, nil).Once()
		gateway.On("MovieById", mock.Anything, "tt2").Return(&model.Movie{Name: "Movie 2"}, nil).Once()
		gateway.On("MovieById", mock.Anything, "tt3").Return(&model.Movie{Name: "Movie 3"}, nil).Once()

		resp, err := svc.GetWeeklySchedule(context.Background(), "2026-03-16")

		assert.Nil(t, err)
		if assert.NotNil(t, resp) {
			assert.Len(t, resp.Schedule, 3)
			assert.Equal(t, 0, resp.Schedule[0].ID)
			assert.Equal(t, "tt2", resp.Schedule[0].Movie.ID)
			assert.Equal(t, 1, resp.Schedule[0].ScreenID)
			assert.Equal(t, 1, resp.Schedule[0].SlotID)
			assert.Equal(t, 300.0, resp.Schedule[0].Cost)
			assert.Equal(t, "tt1", resp.Schedule[1].Movie.ID)
			assert.Equal(t, 280.0, resp.Schedule[1].Cost)
			assert.Equal(t, "tt3", resp.Schedule[2].Movie.ID)
			assert.Equal(t, 320.0, resp.Schedule[2].Cost)
		}
	})
}

func TestAdminScheduleService_DeleteShow(t *testing.T) {
	t.Run("returns bad request for invalid show id", func(t *testing.T) {
		repo := new(MockAdminScheduleRepository)
		gateway := movieGatewayMock.NewMockMovieGateWay(t)
		svc := NewAdminScheduleService(repo, gateway)

		err := svc.DeleteShow(context.Background(), 0)

		assert.NotNil(t, err)
		appErr := err.(*ae.AppError)
		assert.Equal(t, "InvalidShowID", appErr.Code)
	})

	t.Run("deletes show by id", func(t *testing.T) {
		repo := new(MockAdminScheduleRepository)
		gateway := movieGatewayMock.NewMockMovieGateWay(t)
		svc := NewAdminScheduleService(repo, gateway)

		repo.On("DeleteShowByID", mock.Anything, 17).Return(nil).Once()

		err := svc.DeleteShow(context.Background(), 17)

		assert.Nil(t, err)
	})

	t.Run("returns error when tickets are already booked for show", func(t *testing.T) {
		repo := new(MockAdminScheduleRepository)
		gateway := movieGatewayMock.NewMockMovieGateWay(t)
		svc := NewAdminScheduleService(repo, gateway)

		repo.On("DeleteShowByID", mock.Anything, 17).Return(ae.BadRequestError("ShowDeleteNotAllowed", "tickets fot this show already booked, cant delete now", errors.New("has confirmed bookings"))).Once()

		err := svc.DeleteShow(context.Background(), 17)

		assert.NotNil(t, err)
		appErr := err.(*ae.AppError)
		assert.Equal(t, "ShowDeleteNotAllowed", appErr.Code)
		assert.Equal(t, "tickets fot this show already booked, cant delete now", appErr.Message)
	})
}

func TestAdminScheduleService_GetSlots(t *testing.T) {
	t.Run("returns slots from repository", func(t *testing.T) {
		repo := new(MockAdminScheduleRepository)
		gateway := movieGatewayMock.NewMockMovieGateWay(t)
		svc := NewAdminScheduleService(repo, gateway)

		expected := []model.Slot{{Id: 1, Name: "Morning", StartTime: "09:00:00", EndTime: "12:30:00"}}
		repo.On("GetAllSlots", mock.Anything).Return(expected, nil).Once()

		slots, err := svc.GetSlots(context.Background())

		assert.Nil(t, err)
		assert.Equal(t, expected, slots)
	})
}

func TestAdminScheduleService_GetScreens(t *testing.T) {
	t.Run("returns screen options from repository", func(t *testing.T) {
		repo := new(MockAdminScheduleRepository)
		gateway := movieGatewayMock.NewMockMovieGateWay(t)
		svc := NewAdminScheduleService(repo, gateway)

		expected := []model.Screen{{ID: 1, ScreenName: "Screen 1"}, {ID: 2, ScreenName: "Screen 2"}}
		repo.On("GetAllScreens", mock.Anything).Return(expected, nil).Once()

		screens, err := svc.GetScreens(context.Background())

		assert.Nil(t, err)
		assert.Equal(t, 2, len(screens))
		assert.Equal(t, 1, screens[0].ID)
		assert.Equal(t, "Screen 1", screens[0].Name)
		assert.Equal(t, 2, screens[1].ID)
		assert.Equal(t, "Screen 2", screens[1].Name)
	})
}

func TestAdminScheduleService_ScheduleShow(t *testing.T) {
	t.Run("returns bad request when date is invalid", func(t *testing.T) {
		repo := new(MockAdminScheduleRepository)
		gateway := movieGatewayMock.NewMockMovieGateWay(t)
		svc := NewAdminScheduleService(repo, gateway)

		resp, err := svc.ScheduleShow(context.Background(), request.ScheduleShowRequest{Date: "16-03-2026"})

		assert.Nil(t, resp)
		assert.NotNil(t, err)
		appErr := err.(*ae.AppError)
		assert.Equal(t, "InvalidDate", appErr.Code)
	})

	t.Run("returns bad request when slot is occupied", func(t *testing.T) {
		repo := new(MockAdminScheduleRepository)
		gateway := movieGatewayMock.NewMockMovieGateWay(t)
		svc := NewAdminScheduleService(repo, gateway)

		req := request.ScheduleShowRequest{MovieID: "tt123", ScreenID: 1, SlotID: 2, Date: "2026-03-16", Cost: 300}
		showDate := time.Date(2026, time.March, 16, 0, 0, 0, 0, time.UTC)

		repo.On("GetSlotByID", mock.Anything, 2).Return(&model.Slot{Id: 2, StartTime: "13:30:00", EndTime: "17:00:00"}, nil).Once()
		gateway.On("MovieById", mock.Anything, "tt123").Return(&model.Movie{Duration: "148 min"}, nil).Once()
		repo.On("IsSlotOccupied", mock.Anything, 1, showDate, 2).Return(true, nil).Once()

		resp, err := svc.ScheduleShow(context.Background(), req)

		assert.Nil(t, resp)
		assert.NotNil(t, err)
		appErr := err.(*ae.AppError)
		assert.Equal(t, "SlotOccupied", appErr.Code)
	})

	t.Run("returns bad request when movie runtime exceeds slot duration", func(t *testing.T) {
		repo := new(MockAdminScheduleRepository)
		gateway := movieGatewayMock.NewMockMovieGateWay(t)
		svc := NewAdminScheduleService(repo, gateway)

		req := request.ScheduleShowRequest{MovieID: "tt123", ScreenID: 1, SlotID: 1, Date: "2026-03-16", Cost: 350}
		showDate := time.Date(2026, time.March, 16, 0, 0, 0, 0, time.UTC)

		repo.On("GetSlotByID", mock.Anything, 1).Return(&model.Slot{Id: 1, StartTime: "09:00:00", EndTime: "12:30:00"}, nil).Once()
		gateway.On("MovieById", mock.Anything, "tt123").Return(&model.Movie{Duration: "300 min"}, nil).Once()
		repo.On("IsSlotOccupied", mock.Anything, 1, showDate, 1).Return(false, nil).Once()

		resp, err := svc.ScheduleShow(context.Background(), req)

		assert.Nil(t, resp)
		assert.NotNil(t, err)
		appErr := err.(*ae.AppError)
		assert.Equal(t, "MovieRuntimeExceedsSlot", appErr.Code)
	})

	t.Run("creates show when runtime fits slot including overnight slot", func(t *testing.T) {
		repo := new(MockAdminScheduleRepository)
		gateway := movieGatewayMock.NewMockMovieGateWay(t)
		svc := NewAdminScheduleService(repo, gateway)

		req := request.ScheduleShowRequest{MovieID: "tt123", ScreenID: 1, SlotID: 4, Date: "2026-03-16", Cost: 420, Rated: "R"}
		showDate := time.Date(2026, time.March, 16, 0, 0, 0, 0, time.UTC)

		repo.On("GetSlotByID", mock.Anything, 4).Return(&model.Slot{Id: 4, StartTime: "22:30:00", EndTime: "02:00:00"}, nil).Once()
		gateway.On("MovieById", mock.Anything, "tt123").Return(&model.Movie{Duration: "180 min"}, nil).Once()
		repo.On("IsSlotOccupied", mock.Anything, 1, showDate, 4).Return(false, nil).Once()
		repo.On("CreateShow", mock.Anything, mock.AnythingOfType("*model.Show")).Run(func(args mock.Arguments) {
			show := args.Get(1).(*model.Show)
			show.Id = 42
		}).Return(nil).Once()

		resp, err := svc.ScheduleShow(context.Background(), req)

		assert.Nil(t, err)
		if assert.NotNil(t, resp) {
			assert.Equal(t, 42, resp.ID)
			assert.Equal(t, "tt123", resp.MovieID)
			assert.Equal(t, "2026-03-16", resp.Date)
			assert.Equal(t, 420.0, resp.Cost)
			assert.Equal(t, "R", resp.Rated)
		}
	})
}

func TestScheduleService_TimeHelpers(t *testing.T) {
	t.Run("slot duration supports overnight slots", func(t *testing.T) {
		minutes, err := getSlotDurationMinutes("22:30:00", "02:00:00")
		assert.Nil(t, err)
		assert.Equal(t, 210, minutes)
	})

	t.Run("parse runtime supports Go duration format", func(t *testing.T) {
		minutes, err := parseMovieRuntimeMinutes("2h28m0s")
		assert.Nil(t, err)
		assert.Equal(t, 148, minutes)
	})

	t.Run("parse runtime supports min format", func(t *testing.T) {
		minutes, err := parseMovieRuntimeMinutes("148 min")
		assert.Nil(t, err)
		assert.Equal(t, 148, minutes)
	})

	t.Run("parse runtime fails for invalid string", func(t *testing.T) {
		minutes, err := parseMovieRuntimeMinutes("runtime unknown")
		assert.Equal(t, 0, minutes)
		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, err) || err.Error() != "")
	})
}

func TestAdminScheduleService_AddScreen(t *testing.T) {
	t.Run("creates screen with generated name when name is not provided", func(t *testing.T) {
		repo := new(MockAdminScheduleRepository)
		gateway := movieGatewayMock.NewMockMovieGateWay(t)
		svc := NewAdminScheduleService(repo, gateway)

		existing := []model.Screen{{ID: 1, ScreenName: "Screen 1"}, {ID: 2, ScreenName: "Screen 2"}}
		repo.On("GetAllScreens", mock.Anything).Return(existing, nil).Once()
		repo.On("CreateScreen", mock.Anything, mock.AnythingOfType("*model.Screen")).Run(func(args mock.Arguments) {
			screen := args.Get(1).(*model.Screen)
			assert.Equal(t, "Screen 3", screen.ScreenName)
			screen.ID = 3
		}).Return(nil).Once()

		resp, err := svc.AddScreen(context.Background(), request.AddScreenRequest{})

		assert.Nil(t, err)
		if assert.NotNil(t, resp) {
			assert.Equal(t, 3, resp.ID)
			assert.Equal(t, "Screen 3", resp.Name)
		}
	})

	t.Run("creates screen with provided name", func(t *testing.T) {
		repo := new(MockAdminScheduleRepository)
		gateway := movieGatewayMock.NewMockMovieGateWay(t)
		svc := NewAdminScheduleService(repo, gateway)

		repo.On("GetAllScreens", mock.Anything).Return([]model.Screen{}, nil).Once()
		repo.On("CreateScreen", mock.Anything, mock.AnythingOfType("*model.Screen")).Run(func(args mock.Arguments) {
			screen := args.Get(1).(*model.Screen)
			assert.Equal(t, "IMAX Prime", screen.ScreenName)
			screen.ID = 10
		}).Return(nil).Once()

		resp, err := svc.AddScreen(context.Background(), request.AddScreenRequest{Name: "IMAX Prime"})

		assert.Nil(t, err)
		if assert.NotNil(t, resp) {
			assert.Equal(t, 10, resp.ID)
			assert.Equal(t, "IMAX Prime", resp.Name)
		}
	})
}
