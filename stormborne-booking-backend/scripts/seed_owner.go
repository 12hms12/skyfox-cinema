package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"

	"skyfox/bookings/model"
	"skyfox/bookings/repository"
	"skyfox/common/middleware/security"
	"skyfox/config"
	ae "skyfox/error"
	"skyfox/bookings/database/connection"

	"golang.org/x/crypto/bcrypt"
)

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	appErr, ok := err.(*ae.AppError)
	if !ok {
		return false
	}
	return appErr.HTTPCode() == http.StatusNotFound
}

func main() {
	var configFile string
	flag.StringVar(&configFile, "configFile", "", "config file path")
	flag.Parse()

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("[OwnerSeed] failed to load config: %v\n", err)
		os.Exit(1)
	}

	handler := connection.NewDBHandler(cfg.Database)
	db := handler.Instance()
	repo := repository.NewCustomerRepository(db)
	ctx := context.Background()

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
		fmt.Printf("[OwnerSeed] failed to hash owner password: %v\n", err)
		os.Exit(1)
	}
	owner.Password = string(hashedPassword)

	existingByUsername, err := repo.FindByUsername(ctx, owner.Username)
	if err == nil && existingByUsername != nil {
		existingByUsername.UserRole = string(security.OWNER)
		if updateErr := repo.Update(ctx, existingByUsername); updateErr != nil {
			fmt.Printf("[OwnerSeed] failed to update existing owner role by username: %v\n", updateErr)
			os.Exit(1)
		}
		fmt.Printf("[OwnerSeed] owner already exists by username (%s).\n", owner.Username)
		return
	}
	if err != nil && !isNotFound(err) {
		fmt.Printf("[OwnerSeed] failed to lookup owner by username: %v\n", err)
		os.Exit(1)
	}

	existingByEmail, err := repo.FindByEmail(ctx, owner.Email)
	if err == nil && existingByEmail != nil {
		existingByEmail.UserRole = string(security.OWNER)
		if updateErr := repo.Update(ctx, existingByEmail); updateErr != nil {
			fmt.Printf("[OwnerSeed] failed to update existing owner role by email: %v\n", updateErr)
			os.Exit(1)
		}
		fmt.Printf("[OwnerSeed] owner already exists by email (%s).\n", owner.Email)
		return
	}
	if err != nil && !isNotFound(err) {
		fmt.Printf("[OwnerSeed] failed to lookup owner by email: %v\n", err)
		os.Exit(1)
	}

	if err := repo.Create(ctx, &owner); err != nil {
		fmt.Printf("[OwnerSeed] failed to create owner: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[OwnerSeed] owner user created successfully: username=%s, email=%s\n", owner.Username, owner.Email)
}
