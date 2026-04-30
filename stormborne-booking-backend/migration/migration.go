package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"skyfox/bookings/database/connection"
	"skyfox/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	configFileKey     = "configFile"
	defaultConfigFile = ""
	configFileUsage   = "this is config file path"
)

const (
	cutSet   = "file://"
	database = "postgres"
)

func main() {
	var configFile string
	flag.StringVar(&configFile, configFileKey, defaultConfigFile, configFileUsage)
	flag.Parse()

	switch flag.Args()[0] {
	case "up":
		runMigrations(configFile)
	case "down":
		rollBackMigrations(configFile)
	}
}

func runMigrations(configFile string) {
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("error occurred. %v", err)
		return
	}

	m, err := newMigrate(cfg.Database)
	if err != nil {
		fmt.Printf("error occurred: %v", err)
		return
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			fmt.Println("No changes identified for migration")
			return
		}
		fmt.Printf("error occurred: %v", err)
		return
	}
}

func rollBackMigrations(configFile string) {
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("error occurred. %v", err)
	}

	m, _ := newMigrate(cfg.Database)

	if err := m.Down(); err != nil {
		if err == migrate.ErrNoChange {
			fmt.Println("No changes identified for migration")
			return
		}
		fmt.Printf("error occurred: %v", err)
		return
	}
}

func newMigrate(dbCfg config.DbConfig) (*migrate.Migrate, error) {
	directory, err := sourcePath(dbCfg.MigrationPath)
	if err != nil {
		return nil, err
	}
	handler := connection.NewDBHandler(dbCfg)

	sqlDb, err := handler.Instance().SqlDB()
	if err != nil {
		return nil, err
	}

	driver, err := postgres.WithInstance(sqlDb, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(directory, database, driver)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func sourcePath(directory string) (string, error) {
	absPath, err := filepath.Abs(directory)

	if err != nil {
		return "", err
	}

	directory = fmt.Sprintf("%s%s", cutSet, absPath)
	return directory, nil
}
