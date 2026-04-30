package persistence

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"skyfox/bookings/model"
	"skyfox/bookings/repository"
	"skyfox/bookings/service"
	"skyfox/common/middleware/security"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var movieIDs = []string{
	"tt6644200", "tt6857112", "tt7784604", "tt5052448", "tt1396484", "tt5968394",
}

var slotSeed = []struct {
	Name      string
	StartTime string
	EndTime   string
}{
	{"slot1", "09:00:00", "12:30:00"},
	{"slot2", "13:30:00", "17:00:00"},
	{"slot3", "18:00:00", "21:30:00"},
	{"slot4", "22:30:00", "02:00:00"},
}


func randomMovieID(r *rand.Rand) string {
	return movieIDs[r.Intn(len(movieIDs))]
}

func randomPrice(r *rand.Rand) float64 {
	return float64(150+r.Intn(151)) + float64(r.Intn(99))/100
}

func envOrDefault(key, fallback string) string {
    value := os.Getenv(key)
    if value == "" {
        return fallback
    }
    return value
}

func seedOwner(ctx context.Context, repo service.CustomerRepository) {
    owner := model.Customer{
        FirstName:       envOrDefault("OWNER_FIRST_NAME", "Sky"),
        LastName:        envOrDefault("OWNER_LAST_NAME", "Owner"),
        Email:           envOrDefault("OWNER_EMAIL", "owner@skyfox.com"),
        PhoneNumber:     envOrDefault("OWNER_PHONE_NUMBER", "9999999999"),
        CountryCode:     envOrDefault("OWNER_COUNTRY_CODE", "+91"),
        Age:             35,
        Gender:          model.MALE,
        Username:        envOrDefault("OWNER_USERNAME", "skyfox-owner"),
        UserRole:        string(security.OWNER),
        IsEmailVerified: true,
        IsPhoneVerified: true,
    }

    plainPassword := envOrDefault("OWNER_PASSWORD", "Owner@123")
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
    if err != nil {
        fmt.Printf("[Seed] Failed to hash owner password: %v\n", err)
        return
    }
    owner.Password = string(hashedPassword)

    existing, err := repo.FindByUsername(ctx, owner.Username)
    if err == nil && existing != nil {
        fmt.Printf("[Seed] Owner already exists by username (%s). Updating role.\n", owner.Username)
        existing.UserRole = string(security.OWNER)
        _ = repo.Update(ctx, existing)
        return
    }

    existing, err = repo.FindByEmail(ctx, owner.Email)
    if err == nil && existing != nil {
        fmt.Printf("[Seed] Owner already exists by email (%s). Updating role.\n", owner.Email)
        existing.UserRole = string(security.OWNER)
        _ = repo.Update(ctx, existing)
        return
    }

    if err := repo.Create(ctx, &owner); err != nil {
        fmt.Printf("[Seed] Failed to create owner: %v\n", err)
    } else {
        fmt.Printf("[Seed] Owner created: %s\n", owner.Username)
    }
}


func seedUsers(ctx context.Context, userRepo service.CustomerRepository) {
	seedOwner(ctx,userRepo)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("foobar"), bcrypt.DefaultCost)

	candidates := []model.Customer{
		{
			FirstName: "user", LastName: "one", Username: "seed-user-1",
			Password: string(hashedPassword), Email: "userone@skyfox.com",
			PhoneNumber: "1234567890", CountryCode: "+91",
			Gender: model.MALE, Age: 20, UserRole: string(security.ADMIN),
		},
		{
			FirstName: "user", LastName: "two", Username: "seed-user-2",
			Password: string(hashedPassword), Email: "usertwo@skyfox.com",
			PhoneNumber: "1234567891", CountryCode: "+91",
			Gender: model.MALE, Age: 20, UserRole: string(security.ADMIN),
		},
	}

	for _, u := range candidates {
		if _, err := userRepo.FindByUsername(ctx, u.Username); err != nil {
			userCopy := u
			userRepo.Create(ctx, &userCopy)
		}
	}
}

func seedScreen(ctx context.Context, seatRepo repository.SeatRepository, showRepo service.ShowRepository) (*model.Screen, error) {
	screen, err := seatRepo.FindScreenByName(ctx, "Screen 1")
	if err != nil || screen == nil {
		screen = &model.Screen{ScreenName: "Screen 1"}
		if err := showRepo.CreateScreen(ctx, screen); err != nil {
			fmt.Println("[Seed] Failed to create screen:", err)
			return nil, err
		}
		fmt.Println("[Seed] Screen created:", screen.ID)
	} else {
		fmt.Println("[Seed] Screen already exists, ID:", screen.ID)
	}
	return screen, nil
}



func seedSeats(ctx context.Context, seatRepo repository.SeatRepository, screenID int) ([]model.Seat, error) {
	seats, err := seatRepo.FindSeatsByScreenID(ctx, screenID)
	if err != nil || len(seats) == 0 {
		return nil, fmt.Errorf("no seats found for screen %d", screenID)
	}
	fmt.Printf("[Seed] Found %d seats for screen %d\n", len(seats), screenID)
	return seats, nil
}

func seedSlots(ctx context.Context, db *gorm.DB) ([]model.Slot, error) {
	for _, sd := range slotSeed {
		if err := db.WithContext(ctx).Exec(
			`INSERT INTO slot (name, start_time, end_time)
			 VALUES (?, ?, ?) ON CONFLICT (name) DO NOTHING`,
			sd.Name, sd.StartTime, sd.EndTime,
		).Error; err != nil {
			return nil, fmt.Errorf("[Seed] Failed to upsert slot %s: %w", sd.Name, err)
		}
	}
	fmt.Println("[Seed] Slots seeded")

	var slots []model.Slot
	if err := db.WithContext(ctx).Raw(`SELECT id, name, start_time, end_time FROM slot`).
		Scan(&slots).Error; err != nil {
		return nil, fmt.Errorf("[Seed] Failed to query slots: %w", err)
	}
	return slots, nil
}

func seedShows(
	ctx context.Context,
	db *gorm.DB,
	screen *model.Screen,
	slots []model.Slot,
	startDate time.Time,
	totalDays int,
) error {
	if totalDays <= 0 {
		totalDays = 5
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	created, skipped := 0, 0

	for day := 0; day < totalDays; day++ {
		date := startDate.AddDate(0, 0, day)
		for _, slot := range slots {
			tx := db.WithContext(ctx).Exec(
				`INSERT INTO show (screen_id, movie_id, date, slot_id, cost)
				 VALUES (?, ?, ?, ?, ?) ON CONFLICT (screen_id, date, slot_id) DO NOTHING`,
				screen.ID, randomMovieID(r), date.Format("2006-01-02"), slot.Id, randomPrice(r),
			)
			if tx.Error != nil {
				return fmt.Errorf("[Seed] Failed to insert show date=%s slot=%d: %w",
					date.Format("2006-01-02"), slot.Id, tx.Error)
			}
			if tx.RowsAffected == 0 {
				skipped++
			} else {
				created++
			}
		}
	}
	fmt.Printf("[Seed] Shows seeded — created: %d, skipped: %d\n", created, skipped)
	return nil
}

func seedShowPricing(ctx context.Context, seatRepo repository.SeatRepository, show model.Show) error {
	if existing, _ := seatRepo.FindPricingByShowID(ctx, show.Id); len(existing) > 0 {
		return nil
	}
	for _, p := range []model.ShowPricing{
		{ShowID: show.Id, SeatType: "CLASSIC", Price: show.Cost},
		{ShowID: show.Id, SeatType: "CLUB", Price: show.Cost + 50},
		{ShowID: show.Id, SeatType: "RECLINER", Price: show.Cost + 100},
	} {
		pCopy := p
		if err := seatRepo.CreateShowPricing(ctx, &pCopy); err != nil {
			fmt.Printf("[Seed] Failed to create pricing for show %d type %s: %v\n", show.Id, p.SeatType, err)
			return err
		}
	}
	return nil
}

func seedShowSeatStatuses(ctx context.Context, seatRepo repository.SeatRepository, show model.Show, seats []model.Seat) bool {
	if existing, _ := seatRepo.FindSeatStatusByShowID(ctx, show.Id); len(existing) > 0 {
		return false
	}
	for _, seat := range seats {
		status := "AVAILABLE"
		if err := seatRepo.CreateShowSeatStatus(ctx, &model.ShowSeatStatus{
			ShowID: show.Id, SeatID: seat.ID, Status: status,
		}); err != nil {
			fmt.Printf("[Seed] Failed to create seat status show %d seat %d: %v\n", show.Id, seat.ID, err)
			return false
		}
	}
	return true
}

func logTestEndpoints(shows []model.Show) {
	var weekdayID, weekendID int
	for _, show := range shows {
		if weekdayID != 0 && weekendID != 0 {
			break
		}
		day := show.Date.Weekday()
		if (day == 0 || day == 6) && weekendID == 0 {
			weekendID = show.Id
		} else if weekdayID == 0 {
			weekdayID = show.Id
		}
	}
	if weekdayID != 0 {
		fmt.Printf("[Seed] Weekday show (no surcharge) -> GET /shows/%d/seats\n", weekdayID)
	}
	if weekendID != 0 {
		fmt.Printf("[Seed] Weekend show (+50)          -> GET /shows/%d/seats\n", weekendID)
	} else {
		fmt.Println("[Seed] No weekend show found — test any show ID from /show table")
	}
}

// SeedAll populates the database with users, a screen, seats, slots, shows,
// show pricing, and seat statuses. Every step is idempotent; re-running
// against an already-populated database is safe.
//
//   - db        – *gorm.DB used for slot and show seeding (raw SQL)
//   - userRepo  – repository for OnlineCustomer records
//   - seatRepo  – repository for all cinema-seating entities
//   - seedingConfig – first date for show seeding; defaults to today when zero; number of days to seed shows for; defaults to 21

func SeedAll(
	db *gorm.DB,
	userRepo service.CustomerRepository,
	seatRepo repository.SeatRepository,
	showRepo service.ShowRepository,
	seedingConfig SeedingConfig,
) {
	ctx := context.Background()
	startDate := seedingConfig.startDate
	totalDays := seedingConfig.totalDays
	if startDate.IsZero() {
		startDate = time.Now().Truncate(24 * time.Hour)
	}

	seedUsers(ctx, userRepo)

	screen, err := seedScreen(ctx, seatRepo, showRepo) 
	if err != nil {
		return
	}

	seats, err := seedSeats(ctx, seatRepo, screen.ID) 
	if err != nil {
		return
	}
	
	slots, err := seedSlots(ctx, db)
	if err != nil {
		fmt.Println("[Seed] Failed to seed slots:", err)
		return
	}

	if err := seedShows(ctx, db, screen, slots, startDate, totalDays); err != nil {
		fmt.Println("[Seed] Failed to seed shows:", err)
		return
	}

	shows, err := seatRepo.FindShowsByScreenID(ctx, screen.ID)
	if err != nil || len(shows) == 0 {
		fmt.Println("[Seed] No shows found for screen, cannot proceed")
		return
	}
	fmt.Printf("[Seed] Found %d shows for screen %d\n", len(shows), screen.ID)

	seeded, skipped := 0, 0
	for _, show := range shows {
		if err := seedShowPricing(ctx, seatRepo, show); err != nil {
			continue
		}
		if seedShowSeatStatuses(ctx, seatRepo, show, seats) {
			seeded++
		} else {
			skipped++
		}
	}
	fmt.Printf("[Seed] Seat statuses — seeded: %d, skipped: %d\n", seeded, skipped)

	logTestEndpoints(shows)
}