package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"skyfox/bookings/common"
	"skyfox/bookings/dto/response"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSeatService struct {
	mock.Mock
}

func (m *MockSeatService) GetSeatStatus(ctx context.Context, showID uint) (*response.SeatStatusResponse, error) {
	args := m.Called(ctx, showID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*response.SeatStatusResponse), args.Error(1)
}

func TestGetSeatStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		showId      string
		featureFlag bool
		setup       func(svc *MockSeatService)
		wantStatus  int
	}{
		{
			name:        "should return 404 when Feature_SeatMap is disabled",
			showId:      "1",
			featureFlag: false,
			setup:       func(svc *MockSeatService) {},
			wantStatus:  http.StatusNotFound,
		},
		{
			name:        "should return 400 when showId is not a valid uint",
			showId:      "abc",
			featureFlag: true,
			setup:       func(svc *MockSeatService) {},
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "should return 500 when service returns an error",
			showId:      "1",
			featureFlag: true,
			setup: func(svc *MockSeatService) {
				svc.On("GetSeatStatus", mock.Anything, uint(1)).Return(nil, errors.New("show not found"))
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:        "should return 200 with seat status response when service succeeds",
			showId:      "1",
			featureFlag: true,
			setup: func(svc *MockSeatService) {
				svc.On("GetSeatStatus", mock.Anything, uint(1)).Return(&response.SeatStatusResponse{
					ShowID:   1,
					ShowDate: "2025-03-15",
					Seats: []response.SeatResponse{
						{SeatID: 1, Label: "A1", Row: "A", Column: 1, SeatType: "REGULAR", Status: "available", BasePrice: 200.0, IsWeekend: false},
					},
				}, nil)
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := common.Feature_SeatMap
			common.Feature_SeatMap = tt.featureFlag
			defer func() { common.Feature_SeatMap = original }()

			svc := new(MockSeatService)
			tt.setup(svc)

			ctrl := NewSeatController(svc)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodGet, "/shows/"+tt.showId+"/seats", nil)
			c.Params = gin.Params{{Key: "showId", Value: tt.showId}}

			ctrl.GetSeatStatus(c)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var body response.SeatStatusResponse
				json.Unmarshal(w.Body.Bytes(), &body)
				assert.Equal(t, uint(1), body.ShowID)
				assert.Equal(t, "2025-03-15", body.ShowDate)
				assert.Len(t, body.Seats, 1)
				assert.Equal(t, "A1", body.Seats[0].Label)
			}

			if tt.wantStatus == http.StatusNotFound {
				var body map[string]string
				json.Unmarshal(w.Body.Bytes(), &body)
				assert.Equal(t, "feature not available", body["error"])
			}

			if tt.wantStatus == http.StatusInternalServerError {
				var body map[string]string
				json.Unmarshal(w.Body.Bytes(), &body)
				assert.Equal(t, "show not found", body["error"])
			}

			svc.AssertExpectations(t)
		})
	}
}
