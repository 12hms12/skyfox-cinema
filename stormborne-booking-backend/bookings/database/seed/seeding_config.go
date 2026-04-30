package persistence

import "time"

type SeedingConfig struct {
	startDate time.Time
	totalDays int
}