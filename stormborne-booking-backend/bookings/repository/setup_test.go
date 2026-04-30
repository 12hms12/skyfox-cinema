package repository_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"skyfox/common/logger"
	"skyfox/common/middleware/validator"
	"skyfox/config"
	db "skyfox/integration_test/db"
	"testing"
	"time"

	"github.com/appleboy/gofight/v2"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func TestMain(m *testing.M) {
	// setup
	db.SetupTestContainerEnv()
	ctx := context.Background()
	cfg, err := config.LoadConfig(filepath.Join("./../../integration_test", "conf", "test.yml"))
	if err != nil {
		fmt.Printf("unable to load test config. error: %v", err)
	}

	var c *db.DatabaseContainer
	if ci_cd := os.Getenv("RUNNING_ON_CI"); ci_cd != "true" {
		cfg.Database.Host = "localhost"
		c = setupTestContainerDB(ctx, cfg.Database)
		cfg.Database.Port = c.MappedPort.Int()
	}
	initDB(cfg.Database)
	logger.InitAppLogger(cfg.Logger)

	// test execution
	code := m.Run()

	// teardown
	if c != nil && c.Container != nil {
		err := c.Container.Terminate(ctx)
		if err != nil {
			fmt.Printf("error occurred while terminating the test container, %v", err)
			code = 1
		}
	}

	os.Exit(code)
}

func setupTestContainerDB(ctx context.Context, cfg config.DbConfig) *db.DatabaseContainer {

	var container *db.DatabaseContainer
	var err error

	container, err = db.CreateTestContainer(ctx, cfg)
	if err != nil {
		fmt.Printf("%v", err)
		panic("failed to start testdb testcontainer")
	}

	return container
}

func getEngine() (*gin.Engine, *gofight.RequestConfig) {
	_, err := config.LoadConfig(filepath.Join("./", "conf", "test.yml"))
	if err != nil {
		fmt.Printf("unable to load test config. error: %v", err)
	}

	gin.SetMode(gin.TestMode)
	engine := gin.Default()
	binding.Validator = new(validator.DtoValidator)
	engine.Use(ginzap.Ginzap(logger.GetLogger(), time.RFC3339, true))
	engine.Use(ginzap.RecoveryWithZap(logger.GetLogger(), true))

	request := gofight.New()
	return engine, request
}

func initDB(cfg config.DbConfig) {
	testDB := db.InitDB(cfg)
	testDB.Seed()
}
