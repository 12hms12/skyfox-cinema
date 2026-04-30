
package service

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"skyfox/bookings/dto/request"
	"skyfox/bookings/dto/response"
	"skyfox/bookings/model"
	ae "skyfox/error"
	movieservice "skyfox/movieservice/movie_gateway"
)

type AdminScheduleRepository interface {
	GetShowsInRange(ctx context.Context, startDate time.Time, endDate time.Time) ([]model.Show, error)
	GetAllSlots(ctx context.Context) ([]model.Slot, error)
	GetAllScreens(ctx context.Context) ([]model.Screen, error)
	GetSlotByID(ctx context.Context, slotID int) (*model.Slot, error)
	IsSlotOccupied(ctx context.Context, screenID int, date time.Time, slotID int) (bool, error)
	CreateScreen(ctx context.Context, screen *model.Screen) error
	CreateShow(ctx context.Context, show *model.Show) error
	DeleteShowByID(ctx context.Context, showID int) error
}

type adminScheduleService struct {
	repo         AdminScheduleRepository
	movieGateway movieservice.MovieGateWay
}

func NewAdminScheduleService(repo AdminScheduleRepository, movieGateway movieservice.MovieGateWay) *adminScheduleService {
	return &adminScheduleService{
		repo:         repo,
		movieGateway: movieGateway,
	}
}

func (s *adminScheduleService) GetWeeklySchedule(ctx context.Context, startDate string) (*response.WeeklyScheduleResponse, error) {
	dateVariable:="2006-01-02"
	parsedStartDate, err := time.Parse(dateVariable, startDate)
	if err != nil {
		return nil, ae.BadRequestError("InvalidStartDate", "startDate must be in YYYY-MM-DD format", err)
	}

	parsedStartDate = parsedStartDate.UTC()
	endDate := parsedStartDate.AddDate(0, 0, 7)

	shows, err := s.repo.GetShowsInRange(ctx, parsedStartDate, endDate)
	if err != nil {
		return nil, err
	}

	schedule := make([]response.ScheduleItem, 0, len(shows))
	for _, show := range shows {
		movie := response.ScheduledMovieBrief{ID: show.MovieId, Name: ""}
		movieDetails, movieErr := s.movieGateway.MovieById(ctx, show.MovieId)
		if movieErr == nil {
			movie.Name = movieDetails.Name
		}

		schedule = append(schedule, response.ScheduleItem{
			ID:       show.Id,
			Date:     show.Date.Format(dateVariable),
			ScreenID: show.ScreenID,
			SlotID:   show.SlotId,
			Cost:     show.Cost,
			Movie:    movie,
		})
	}

	sort.Slice(schedule, func(i, j int) bool {
		if schedule[i].Date != schedule[j].Date {
			return schedule[i].Date < schedule[j].Date
		}
		if schedule[i].ScreenID != schedule[j].ScreenID {
			return schedule[i].ScreenID < schedule[j].ScreenID
		}
		return schedule[i].SlotID < schedule[j].SlotID
	})

	return &response.WeeklyScheduleResponse{Schedule: schedule}, nil
}

func (s *adminScheduleService) GetSlots(ctx context.Context) ([]model.Slot, error) {
	return s.repo.GetAllSlots(ctx)
}

func (s *adminScheduleService) GetScreens(ctx context.Context) ([]response.ScreenOption, error) {
	screens, err := s.repo.GetAllScreens(ctx)
	if err != nil {
		return nil, err
	}

	options := make([]response.ScreenOption, 0, len(screens))
	for _, screen := range screens {
		options = append(options, response.ScreenOption{
			ID:   screen.ID,
			Name: screen.ScreenName,
		})
	}

	return options, nil
}

func (s *adminScheduleService) ScheduleShow(ctx context.Context, req request.ScheduleShowRequest) (*response.ScheduleShowResponse, error) {
	dateVariable:="2006-01-02"
	showDate, err := time.Parse(dateVariable, req.Date)
	if err != nil {
		return nil, ae.BadRequestError("InvalidDate", "date must be in YYYY-MM-DD format", err)
	}
	showDate = showDate.UTC()

	slot, err := s.repo.GetSlotByID(ctx, req.SlotID)
	if err != nil {
		return nil, err
	}

	movie, err := s.movieGateway.MovieById(ctx, req.MovieID)
	if err != nil {
		return nil, err
	}

	occupied, err := s.repo.IsSlotOccupied(ctx, req.ScreenID, showDate, req.SlotID)
	if err != nil {
		return nil, err
	}
	if occupied {
		return nil, ae.BadRequestError("SlotOccupied", "selected slot is already occupied", fmt.Errorf("slot %d is occupied", req.SlotID))
	}

	movieRuntimeMinutes, err := parseMovieRuntimeMinutes(movie.Duration)
	if err != nil {
		return nil, ae.BadRequestError("InvalidMovieRuntime", "unable to parse movie runtime", err)
	}

	slotDurationMinutes, err := getSlotDurationMinutes(slot.StartTime, slot.EndTime)
	if err != nil {
		return nil, ae.UnProcessableError("InvalidSlotConfig", "invalid slot start/end time", err)
	}

	if movieRuntimeMinutes > slotDurationMinutes {
		return nil, ae.BadRequestError(
			"MovieRuntimeExceedsSlot",
			"movie runtime exceeds selected slot duration",
			fmt.Errorf("runtime=%d slotDuration=%d", movieRuntimeMinutes, slotDurationMinutes),
		)
	}

	show := &model.Show{
		MovieId:  req.MovieID,
		ScreenID: req.ScreenID,
		SlotId:   req.SlotID,
		Date:     showDate,
		Cost:     req.Cost,
	}

	if err := s.repo.CreateShow(ctx, show); err != nil {
		return nil, err
	}

	return &response.ScheduleShowResponse{
		ID:       show.Id,
		MovieID:  show.MovieId,
		ScreenID: show.ScreenID,
		SlotID:   show.SlotId,
		Date:     show.Date.Format(dateVariable),
		Cost:     show.Cost,
		Rated:    req.Rated,
	}, nil
}

func getSlotDurationMinutes(startTime string, endTime string) (int, error) {
	timeVariable:="15:04:05"
	start, err := time.Parse(timeVariable, startTime)
	if err != nil {
		return 0, err
	}
	end, err := time.Parse(timeVariable,endTime)
	if err != nil {
		return 0, err
	}

	if !end.After(start) {
		end = end.Add(24 * time.Hour)
	}

	duration := end.Sub(start)
	return int(duration.Minutes()), nil
}

func parseMovieRuntimeMinutes(runtime string) (int, error) {
	runtime = strings.TrimSpace(runtime)
	if runtime == "" {
		return 0, fmt.Errorf("runtime is empty")
	}

	if parsedDuration, err := time.ParseDuration(runtime); err == nil {
		return int(parsedDuration.Minutes()), nil
	}

	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(runtime, -1)
	if len(matches) == 0 {
		return 0, fmt.Errorf("runtime format not supported: %s", runtime)
	}

	minutes, err := strconv.Atoi(matches[0])
	if err != nil {
		return 0, err
	}
	return minutes, nil
}

func (s *adminScheduleService) DeleteShow(ctx context.Context, showID int) error {
	if showID <= 0 {
		return ae.BadRequestError("InvalidShowID", "showId must be a positive integer", fmt.Errorf("invalid showId: %d", showID))
	}

	return s.repo.DeleteShowByID(ctx, showID)
}


func (s *adminScheduleService) AddScreen(ctx context.Context, req request.AddScreenRequest) (*response.ScreenOption, error) {
	screens, err := s.repo.GetAllScreens(ctx)
	if err != nil {
		return nil, err
	}

	providedName := strings.TrimSpace(req.Name)
	if providedName == "" {
		nextScreenNumber := 1
		re := regexp.MustCompile(`(?i)^screen\s+(\d+)$`)
		for _, existing := range screens {
			if existing.ID >= nextScreenNumber {
				nextScreenNumber = existing.ID + 1
			}

			matches := re.FindStringSubmatch(strings.TrimSpace(existing.ScreenName))
			if len(matches) == 2 {
				if parsedNumber, parseErr := strconv.Atoi(matches[1]); parseErr == nil && parsedNumber >= nextScreenNumber {
					nextScreenNumber = parsedNumber + 1
				}
			}
		}

		providedName = fmt.Sprintf("Screen %d", nextScreenNumber)
	}

	newScreen := &model.Screen{ScreenName: providedName}
	if err := s.repo.CreateScreen(ctx, newScreen); err != nil {
		return nil, err
	}

	return &response.ScreenOption{ID: newScreen.ID, Name: newScreen.ScreenName}, nil
}
