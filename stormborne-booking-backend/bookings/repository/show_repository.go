package repository

import (
	"context"
	"fmt"
	"skyfox/bookings/constants"
	"skyfox/bookings/database/common"
	"skyfox/common/logger"
	movieservice "skyfox/movieservice/movie_gateway"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"skyfox/bookings/model"
	ae "skyfox/error"

	"gorm.io/gorm"
)

type showRepository struct {
	*common.BaseDB
	seatRepo     SeatRepository
	movieGateway movieservice.MovieGateWay
}

type RevenueQuery struct {
	StartDate *time.Time
	EndDate   *time.Time
	MovieId   string
	Genre     string
	ShowTime  string
}

func NewShowRepository(db *common.BaseDB, movieGateway movieservice.MovieGateWay, seatRepo SeatRepository) *showRepository {
	return &showRepository{
		BaseDB:       db,
		movieGateway: movieGateway,
		seatRepo:     seatRepo,
	}
}

func (repo showRepository) GetAllShowsOn(ctx context.Context, date string) ([]model.Show, error) {
	var shows []model.Show = make([]model.Show, 0)

	db, cancel := repo.WithContext(ctx)
	defer cancel()

	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, ae.InternalServerError("InvalidDate", "invalid date format", err)
	}

	startOfDay := parsedDate
	endOfDay := parsedDate.Add(24 * time.Hour)

	err = db.Model(&model.Show{}).
		Preload("Slot").
		Where("date >= ? AND date < ?", startOfDay, endOfDay).
		Find(&shows).Error

	if err != nil {
		if err == context.DeadlineExceeded {
			return nil, ae.InternalServerError("InternalServerError", "query could not be processed", err)
		}
		return nil, err
	}

	return shows, nil
}

func (repo showRepository) FindById(ctx context.Context, id int) (model.Show, error) {
	var show model.Show

	db, cancel := repo.WithContext(ctx)
	defer cancel()

	result := db.Model(&model.Show{}).Preload("Slot").Where("id=?", id).First(&show)

	if result.Error != nil {
		return model.Show{}, ae.NotFoundError("ShowNotFound", fmt.Sprintf("Show not found for id : %d", id), result.Error)
	}
	return show, nil
}

func (repo showRepository) GetShowsInRange(ctx context.Context, startDate time.Time, endDate time.Time) ([]model.Show, error) {
	var shows []model.Show = make([]model.Show, 0)

	db, cancel := repo.WithContext(ctx)
	defer cancel()

	err := db.Model(&model.Show{}).
		Preload("Slot").
		Where("date >= ? AND date < ?", startDate, endDate).
		Find(&shows).Error

	if err != nil {
		return nil, ae.InternalServerError("InternalServerError", "query could not be processed", err)
	}

	return shows, nil
}

func (repo showRepository) GetSlotByID(ctx context.Context, slotID int) (*model.Slot, error) {
	var slot model.Slot

	db, cancel := repo.WithContext(ctx)
	defer cancel()

	result := db.Model(&model.Slot{}).Where("id = ?", slotID).First(&slot)
	if result.Error != nil {
		return nil, ae.NotFoundError("SlotNotFound", fmt.Sprintf("slot not found for id : %d", slotID), result.Error)
	}

	return &slot, nil
}

func (repo showRepository) GetAllSlots(ctx context.Context) ([]model.Slot, error) {
	var slots []model.Slot

	db, cancel := repo.WithContext(ctx)
	defer cancel()

	result := db.Model(&model.Slot{}).Order("id ASC").Find(&slots)
	if result.Error != nil {
		return nil, ae.InternalServerError("SlotFetchFailed", "failed to fetch slots", result.Error)
	}

	return slots, nil
}

func (repo showRepository) GetAllScreens(ctx context.Context) ([]model.Screen, error) {
	var screens []model.Screen

	db, cancel := repo.WithContext(ctx)
	defer cancel()

	result := db.Model(&model.Screen{}).Order("id ASC").Find(&screens)
	if result.Error != nil {
		return nil, ae.InternalServerError("ScreenFetchFailed", "failed to fetch screens", result.Error)
	}

	return screens, nil
}

func (repo showRepository) IsSlotOccupied(ctx context.Context, screenID int, date time.Time, slotID int) (bool, error) {
	var count int64

	db, cancel := repo.WithContext(ctx)
	defer cancel()

	result := db.Model(&model.Show{}).
		Where("screen_id = ? AND date = ? AND slot_id = ?", screenID, date, slotID).
		Count(&count)

	if result.Error != nil {
		return false, ae.InternalServerError("ShowFetchFailed", "failed to verify slot occupancy", result.Error)
	}

	return count > 0, nil
}

func (repo showRepository) CreateShow(ctx context.Context, show *model.Show) error {
	db, cancel := repo.WithContext(ctx)
	defer cancel()

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(show).Error; err != nil {
			if strings.Contains(err.Error(), constants.Pg_duplicate_error) {
				return ae.BadRequestError("SlotOccupied", "selected slot is already occupied", err)
			}
			return ae.UnProcessableError("ShowCreationFailed", "Failed to create show", err)
		}

		seats, err := repo.seatRepo.FindSeatsByScreenID(ctx, show.ScreenID)
		if err != nil {
			return ae.UnProcessableError("SeatFetchFailed", "Failed to fetch seats", err)
		}

		for _, seat := range seats {
			if err := tx.Create(&model.ShowSeatStatus{
				ShowID: show.Id,
				SeatID: seat.ID,
				Status: "AVAILABLE",
			}).Error; err != nil {
				return ae.UnProcessableError("SeatStatusCreationFailed",
					fmt.Sprintf("Failed for show %d seat %d", show.Id, seat.ID), err)
			}
		}

		for _, p := range []model.ShowPricing{
			{ShowID: show.Id, SeatType: "CLASSIC", Price: show.Cost},
			{ShowID: show.Id, SeatType: "CLUB", Price: show.Cost + 50},
			{ShowID: show.Id, SeatType: "RECLINER", Price: show.Cost + 100},
		} {
			if err := tx.Create(&p).Error; err != nil {
				return ae.UnProcessableError("PricingCreationFailed",
					fmt.Sprintf("Failed to create pricing for show %d type %s", show.Id, p.SeatType), err)
			}
		}

		return nil
	})
}

func (repo showRepository) DeleteShowByID(ctx context.Context, showID int) error {
	db, cancel := repo.WithContext(ctx)
	defer cancel()
	var bookedTicketCount int64
	countResult := db.Model(&model.Booking{}).
		Where("show_id = ? AND status = ?", showID, "CONFIRMED").
		Count(&bookedTicketCount)
	if countResult.Error != nil {
		return ae.UnProcessableError("ShowDeleteCheckFailed", "failed to check show bookings before delete", countResult.Error)
	}

	if bookedTicketCount > 0 {
		return ae.BadRequestError(
			"ShowDeleteNotAllowed",
			"tickets fot this show already booked, cant delete now",
			fmt.Errorf("show %d has %d confirmed bookings", showID, bookedTicketCount),
		)
	}
	
	result := db.Where("id = ?", showID).Delete(&model.Show{})
	if result.Error != nil {
		return ae.UnProcessableError("ShowDeleteFailed", "failed to delete show", result.Error)
	}

	if result.RowsAffected == 0 {
		return ae.NotFoundError("ShowNotFound", fmt.Sprintf("show not found for id : %d", showID), gorm.ErrRecordNotFound)
	}

	return nil
}

func (repo showRepository) GetAllShowsBy(ctx context.Context, revenueQuery *RevenueQuery) ([]model.Show, error) {
	var shows []model.Show

	db, cancel := repo.WithContext(ctx)
	defer cancel()

	query := db.Joins("Slot").Model(&model.Show{})

	if revenueQuery.MovieId != "" {
		query = query.Where("movie_id = ?", revenueQuery.MovieId)
	}

	if revenueQuery.StartDate != nil && !revenueQuery.StartDate.IsZero() {
		query = query.Where("date >= ?", revenueQuery.StartDate)
	}

	if revenueQuery.EndDate != nil && !revenueQuery.StartDate.IsZero() {
		query = query.Where("date <= ?", revenueQuery.EndDate)
	}

	if revenueQuery.ShowTime != "" {
		query = query.Where("\"Slot\".start_time = ?", revenueQuery.ShowTime)
	}

	result := query.Find(&shows)

	if result.Error != nil {
		return nil, ae.InternalServerError("QueryError", "Database error", result.Error)
	}

	filteredShows := make([]model.Show, 0, len(shows))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, s := range shows {
		wg.Add(1)
		go func(show model.Show) {
			defer wg.Done()
			movie, err := repo.movieGateway.MovieById(ctx, show.MovieId)
			if err != nil {
				logger.Error("unable to fetch movie details")
				return
			}

			isGenreMatch := revenueQuery.Genre == "" ||
				slices.Contains(strings.Split(movie.Genre, ", "), revenueQuery.Genre)

			if isGenreMatch {
				show.Movie = *movie
				mu.Lock()
				filteredShows = append(filteredShows, show)
				mu.Unlock()
			}
		}(s)
	}
	wg.Wait()

	// Sort by Show ID ascending
	sort.Slice(filteredShows, func(i, j int) bool {
		return filteredShows[i].Id < filteredShows[j].Id
	})

	return filteredShows, nil
}

func (repo showRepository) GetScheduledMovieIds(ctx context.Context) ([]string, error) {
	var movieIds []string

	db, cancel := repo.WithContext(ctx)
	defer cancel()

	startOfToday := time.Now().UTC().Truncate(24 * time.Hour)

	err := db.Model(&model.Show{}).
		Distinct("movie_id").
		Where("date >= ?", startOfToday).
		Order("movie_id ASC").
		Pluck("movie_id", &movieIds).Error

	if err != nil {
		return nil, ae.InternalServerError("MovieIdsFetchFailed", "failed to fetch scheduled movie ids", err)
	}

	return movieIds, nil
}

func (repo showRepository) GetShowsByMovieAndDate(ctx context.Context, movieId, date string) ([]model.Show, error) {
	var shows []model.Show

	db, cancel := repo.WithContext(ctx)
	defer cancel()

	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, ae.InternalServerError("InvalidDate", "invalid date format", err)
	}

	startOfDay := parsedDate
	endOfDay := parsedDate.Add(24 * time.Hour)

	err = db.Model(&model.Show{}).
		Preload("Slot").
		Preload("Screen").
		Where("movie_id = ? AND date >= ? AND date < ?", movieId, startOfDay, endOfDay).
		Order("date ASC").
		Find(&shows).Error

	if err != nil {
		return nil, ae.InternalServerError("ShowsFetchFailed", "failed to fetch shows for movie and date", err)
	}

	return shows, nil
}

type ShowDate struct {
	ShowDate string `gorm:"column:show_date"`
}

func (repo showRepository) GetAvailableDatesForMovie(ctx context.Context, movieId string) ([]string, error) {
	db, cancel := repo.WithContext(ctx)
	defer cancel()

	startOfToday := time.Now().UTC().Truncate(24 * time.Hour)

	var results []ShowDate
	err := db.Raw(
		`SELECT DISTINCT TO_CHAR(date AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS show_date FROM "show" WHERE movie_id = ? AND date >= ? ORDER BY show_date ASC`,
		movieId, startOfToday,
	).Scan(&results).Error

	if err != nil {
		return nil, ae.InternalServerError("DatesFetchFailed", "failed to fetch available dates for movie", err)
	}

	dates := make([]string, 0, len(results))
	for _, r := range results {
		dates = append(dates, r.ShowDate)
	}

	return dates, nil
}

func (repo showRepository) CountAvailableSeats(ctx context.Context, showId int) (int64, error) {
	var count int64

	db, cancel := repo.WithContext(ctx)
	defer cancel()

	err := db.Model(&model.ShowSeatStatus{}).
		Where("show_id = ? AND status = 'AVAILABLE'", showId).
		Count(&count).Error

	if err != nil {
		return 0, ae.InternalServerError("CountSeatsFailed", "failed to count available seats", err)
	}

	return count, nil
}

func (repo showRepository) CreateScreen(ctx context.Context, screen *model.Screen) error {
	db, cancel := repo.WithContext(ctx)
	defer cancel()

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(screen).Error; err != nil {
			return ae.UnProcessableError("ScreenCreationFailed", "failed to create screen", err)
		}

		seatLayout := []struct {
					from     int
					to       int
					seatType string
				}{
					{1, 50, "CLASSIC"},
					{51, 80, "CLUB"},
					{81, 100, "RECLINER"},
				}


		seatNum := 1
		for _, l := range seatLayout {
			for i := l.from; i <= l.to; i++ {
				row := ((seatNum - 1) / 10) + 1
				col := ((seatNum - 1) % 10) + 1
				if err := tx.Create(&model.Seat{
					ScreenID:     screen.ID,
					RowNumber:    row,
					ColumnNumber: col,
					SeatType:     l.seatType,
				}).Error; err != nil {
					return ae.UnProcessableError("SeatCreationFailed", "failed to create seats for new screen", err)
				}
				seatNum++
			}
		}
		return nil
	})
}

