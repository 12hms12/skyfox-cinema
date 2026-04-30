package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"skyfox/bookings/dto/request"
	"skyfox/bookings/dto/response"
	"skyfox/bookings/model"
	ae "skyfox/error"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAdminScheduleService struct {
	mock.Mock
}

func (m *MockAdminScheduleService) GetWeeklySchedule(ctx context.Context, startDate string) (*response.WeeklyScheduleResponse, error) {
	args := m.Called(ctx, startDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.WeeklyScheduleResponse), args.Error(1)
}

func (m *MockAdminScheduleService) ScheduleShow(ctx context.Context, req request.ScheduleShowRequest) (*response.ScheduleShowResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ScheduleShowResponse), args.Error(1)
}

func (m *MockAdminScheduleService) DeleteShow(ctx context.Context, showID int) error {
	args := m.Called(ctx, showID)
	return args.Error(0)
}

func (m *MockAdminScheduleService) GetSlots(ctx context.Context) ([]model.Slot, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Slot), args.Error(1)
}

func (m *MockAdminScheduleService) GetScreens(ctx context.Context) ([]response.ScreenOption, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]response.ScreenOption), args.Error(1)
}

func (m *MockAdminScheduleService) AddScreen(ctx context.Context, req request.AddScreenRequest) (*response.ScreenOption, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.ScreenOption), args.Error(1)
}


func TestAdminScheduleController_GetWeeklySchedule(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("returns bad request when startDate missing", func(t *testing.T) {
		svc := new(MockAdminScheduleService)
		ctrl := NewAdminScheduleController(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/schedule/week", nil)

		ctrl.GetWeeklySchedule(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns ok response", func(t *testing.T) {
		svc := new(MockAdminScheduleService)
		ctrl := NewAdminScheduleController(svc)

		expected := &response.WeeklyScheduleResponse{
			Schedule: []response.ScheduleItem{{Date: "2026-03-16", ScreenID: 1, SlotID: 1, Cost: 300, Movie: response.ScheduledMovieBrief{ID: "tt1", Name: "Inception"}}},
		}
		svc.On("GetWeeklySchedule", mock.Anything, "2026-03-16").Return(expected, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/schedule/week?startDate=2026-03-16", nil)

		ctrl.GetWeeklySchedule(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("returns service error", func(t *testing.T) {
		svc := new(MockAdminScheduleService)
		ctrl := NewAdminScheduleController(svc)

		svc.On("GetWeeklySchedule", mock.Anything, "2026-03-16").Return(nil, ae.BadRequestError("InvalidStartDate", "invalid", errors.New("invalid date"))).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/schedule/week?startDate=2026-03-16", nil)

		ctrl.GetWeeklySchedule(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAdminScheduleController_ScheduleShow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("returns bad request on invalid payload", func(t *testing.T) {
		svc := new(MockAdminScheduleService)
		ctrl := NewAdminScheduleController(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/admin/shows", bytes.NewBufferString("{invalid-json"))
		c.Request.Header.Set("Content-Type", "application/json")

		ctrl.ScheduleShow(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns created response", func(t *testing.T) {
		svc := new(MockAdminScheduleService)
		ctrl := NewAdminScheduleController(svc)

		req := request.ScheduleShowRequest{MovieID: "tt123", ScreenID: 1, SlotID: 2, Date: "2026-03-16", Cost: 400, Rated: "R"}
		expected := &response.ScheduleShowResponse{ID: 99, MovieID: "tt123", ScreenID: 1, SlotID: 2, Date: "2026-03-16", Cost: 400, Rated: "R"}
		svc.On("ScheduleShow", mock.Anything, req).Return(expected, nil).Once()

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/admin/shows", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		ctrl.ScheduleShow(c)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("returns service error", func(t *testing.T) {
		svc := new(MockAdminScheduleService)
		ctrl := NewAdminScheduleController(svc)

		req := request.ScheduleShowRequest{MovieID: "tt123", ScreenID: 1, SlotID: 2, Date: "2026-03-16", Cost: 400, Rated: "R"}
		svc.On("ScheduleShow", mock.Anything, req).Return(nil, ae.BadRequestError("SlotOccupied", "occupied", errors.New("occupied"))).Once()

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/admin/shows", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		ctrl.ScheduleShow(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAdminScheduleController_DeleteShow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns bad request on invalid show id", func(t *testing.T) {
		svc := new(MockAdminScheduleService)
		ctrl := NewAdminScheduleController(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "showId", Value: "abc"}}
		c.Request, _ = http.NewRequest(http.MethodDelete, "/admin/shows/abc", nil)

		ctrl.DeleteShow(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns ok response", func(t *testing.T) {
		svc := new(MockAdminScheduleService)
		ctrl := NewAdminScheduleController(svc)
		svc.On("DeleteShow", mock.Anything, 99).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "showId", Value: "99"}}
		c.Request, _ = http.NewRequest(http.MethodDelete, "/admin/shows/99", nil)

		ctrl.DeleteShow(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("returns service error", func(t *testing.T) {
		svc := new(MockAdminScheduleService)
		ctrl := NewAdminScheduleController(svc)
		svc.On("DeleteShow", mock.Anything, 99).Return(ae.NotFoundError("ShowNotFound", "not found", errors.New("not found"))).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "showId", Value: "99"}}
		c.Request, _ = http.NewRequest(http.MethodDelete, "/admin/shows/99", nil)

		ctrl.DeleteShow(c)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestAdminScheduleController_GetSlots(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns ok response", func(t *testing.T) {
		svc := new(MockAdminScheduleService)
		ctrl := NewAdminScheduleController(svc)

		expected := []model.Slot{{Id: 1, Name: "slot-1", StartTime: "09:00:00", EndTime: "12:30:00"}}
		svc.On("GetSlots", mock.Anything).Return(expected, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/slots", nil)

		ctrl.GetSlots(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("returns service error", func(t *testing.T) {
		svc := new(MockAdminScheduleService)
		ctrl := NewAdminScheduleController(svc)

		svc.On("GetSlots", mock.Anything).Return(nil, ae.InternalServerError("SlotFetchFailed", "failed", errors.New("db down"))).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/slots", nil)

		ctrl.GetSlots(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestAdminScheduleController_GetScreens(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns ok response", func(t *testing.T) {
		svc := new(MockAdminScheduleService)
		ctrl := NewAdminScheduleController(svc)

		expected := []response.ScreenOption{{ID: 1, Name: "Screen 1"}}
		svc.On("GetScreens", mock.Anything).Return(expected, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/screens", nil)

		ctrl.GetScreens(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("returns service error", func(t *testing.T) {
		svc := new(MockAdminScheduleService)
		ctrl := NewAdminScheduleController(svc)

		svc.On("GetScreens", mock.Anything).Return(nil, ae.InternalServerError("ScreenFetchFailed", "failed", errors.New("db down"))).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/admin/screens", nil)

		ctrl.GetScreens(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}



func TestAdminScheduleController_AddScreen(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns created response", func(t *testing.T) {
		svc := new(MockAdminScheduleService)
		ctrl := NewAdminScheduleController(svc)

		req := request.AddScreenRequest{}
		expected := &response.ScreenOption{ID: 3, Name: "Screen 3"}
		svc.On("AddScreen", mock.Anything, req).Return(expected, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/admin/screens", nil)

		ctrl.AddScreen(c)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("returns bad request on invalid payload", func(t *testing.T) {
		svc := new(MockAdminScheduleService)
		ctrl := NewAdminScheduleController(svc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/admin/screens", bytes.NewBufferString("{invalid-json"))
		c.Request.Header.Set("Content-Type", "application/json")

		ctrl.AddScreen(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns service error", func(t *testing.T) {
		svc := new(MockAdminScheduleService)
		ctrl := NewAdminScheduleController(svc)

		req := request.AddScreenRequest{}
		svc.On("AddScreen", mock.Anything, req).Return(nil, ae.InternalServerError("ScreenCreationFailed", "failed", errors.New("db down"))).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/admin/screens", nil)

		ctrl.AddScreen(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
