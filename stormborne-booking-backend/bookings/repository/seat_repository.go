package repository

import (
	"context"

	"skyfox/bookings/database/common"
	"skyfox/bookings/model"
	ae "skyfox/error"
)

type SeatRepository interface {
	GetShowByID(ctx context.Context, showID uint) (*model.Show, error)
	GetSeatStatusByShowID(ctx context.Context, showID uint) ([]model.ShowSeatStatus, error)
	GetPricingByShowID(ctx context.Context, showID uint) ([]model.ShowPricing, error)

	FindScreenByName(ctx context.Context, name string) (*model.Screen, error)
	FindSeatsByScreenID(ctx context.Context, screenID int) ([]model.Seat, error)
	CreateSeat(ctx context.Context, seat *model.Seat) error
	FindShowByScreenAndMovie(ctx context.Context, screenID int, movieID string) (*model.Show, error)
	CreateShow(ctx context.Context, show *model.Show) error
	FindPricingByShowID(ctx context.Context, showID int) ([]model.ShowPricing, error)
	CreateShowPricing(ctx context.Context, pricing *model.ShowPricing) error
	FindSeatStatusByShowID(ctx context.Context, showID int) ([]model.ShowSeatStatus, error)
	CreateShowSeatStatus(ctx context.Context, sss *model.ShowSeatStatus) error
	FindShowsByScreenID(ctx context.Context, screenID int) ([]model.Show, error)
}

type seatRepository struct {
	*common.BaseDB
}

func (repo seatRepository) FindShowsByScreenID(ctx context.Context, screenID int) ([]model.Show, error) {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	var shows []model.Show
	result := dbCtx.Where("screen_id = ?", screenID).Find(&shows)
	if result.Error != nil {
		return nil, ae.UnProcessableError("FetchShowsFailed", "Failed to fetch shows", result.Error)
	}
	return shows, nil
}

func NewSeatRepository(db *common.BaseDB) SeatRepository {
	return &seatRepository{BaseDB: db}
}


func (repo seatRepository) GetShowByID(ctx context.Context, showID uint) (*model.Show, error) {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	var show model.Show
	result := dbCtx.First(&show, showID)
	if result.Error != nil {
		return nil, ae.UnProcessableError("ShowNotFound", "Show not found", result.Error)
	}
	return &show, nil
}

func (repo seatRepository) GetSeatStatusByShowID(ctx context.Context, showID uint) ([]model.ShowSeatStatus, error) {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	var seatStatuses []model.ShowSeatStatus
	result := dbCtx.
		Preload("Seat").
		Where("show_id = ?", showID).
		Find(&seatStatuses)

	if result.Error != nil {
		return nil, ae.UnProcessableError("FetchSeatStatusFailed", "Failed to fetch seat statuses", result.Error)
	}
	return seatStatuses, nil
}

func (repo seatRepository) GetPricingByShowID(ctx context.Context, showID uint) ([]model.ShowPricing, error) {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	var pricing []model.ShowPricing
	result := dbCtx.Where("show_id = ?", showID).Find(&pricing)
	if result.Error != nil {
		return nil, ae.UnProcessableError("FetchPricingFailed", "Failed to fetch show pricing", result.Error)
	}
	return pricing, nil
}


func (repo seatRepository) FindScreenByName(ctx context.Context, name string) (*model.Screen, error) {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	var screen model.Screen
	result := dbCtx.Where("screen_name = ?", name).First(&screen)
	if result.Error != nil {
		return nil, nil 
	}
	return &screen, nil
}

func (repo seatRepository) FindSeatsByScreenID(ctx context.Context, screenID int) ([]model.Seat, error) {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	var seats []model.Seat
	result := dbCtx.Where("screen_id = ?", screenID).Find(&seats)
	if result.Error != nil {
		return nil, ae.UnProcessableError("FetchSeatsFailed", "Failed to fetch seats", result.Error)
	}
	return seats, nil
}

func (repo seatRepository) CreateSeat(ctx context.Context, seat *model.Seat) error {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	result := dbCtx.Create(seat)
	if result.Error != nil {
		return ae.UnProcessableError("SeatCreationFailed", "Failed to create seat", result.Error)
	}
	return nil
}

func (repo seatRepository) FindShowByScreenAndMovie(ctx context.Context, screenID int, movieID string) (*model.Show, error) {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	var show model.Show
	result := dbCtx.Where("screen_id = ? AND movie_id = ?", screenID, movieID).First(&show)
	if result.Error != nil {
		return nil, nil 
	}
	return &show, nil
}

func (repo seatRepository) CreateShow(ctx context.Context, show *model.Show) error {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	result := dbCtx.Create(show)
	if result.Error != nil {
		return ae.UnProcessableError("ShowCreationFailed", "Failed to create show", result.Error)
	}
	return nil
}

func (repo seatRepository) FindPricingByShowID(ctx context.Context, showID int) ([]model.ShowPricing, error) {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	var pricing []model.ShowPricing
	result := dbCtx.Where("show_id = ?", showID).Find(&pricing)
	if result.Error != nil {
		return nil, ae.UnProcessableError("FetchPricingFailed", "Failed to fetch pricing", result.Error)
	}
	return pricing, nil
}

func (repo seatRepository) CreateShowPricing(ctx context.Context, pricing *model.ShowPricing) error {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	result := dbCtx.Create(pricing)
	if result.Error != nil {
		return ae.UnProcessableError("PricingCreationFailed", "Failed to create pricing", result.Error)
	}
	return nil
}

func (repo seatRepository) FindSeatStatusByShowID(ctx context.Context, showID int) ([]model.ShowSeatStatus, error) {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	var statuses []model.ShowSeatStatus
	result := dbCtx.Where("show_id = ?", showID).Find(&statuses)
	if result.Error != nil {
		return nil, ae.UnProcessableError("FetchSeatStatusFailed", "Failed to fetch seat statuses", result.Error)
	}
	return statuses, nil
}

func (repo seatRepository) CreateShowSeatStatus(ctx context.Context, sss *model.ShowSeatStatus) error {
	dbCtx, cancel := repo.WithContext(ctx)
	defer cancel()

	result := dbCtx.Create(sss)
	if result.Error != nil {
		return ae.UnProcessableError("SeatStatusCreationFailed", "Failed to create seat status", result.Error)
	}
	return nil
}


