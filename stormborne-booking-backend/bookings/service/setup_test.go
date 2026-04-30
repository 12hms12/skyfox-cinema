package service

import (
	"fmt"
	"os"
	"path/filepath"
	"skyfox/common/logger"
	"skyfox/config"
	"testing"
)

func TestMain(m *testing.M) {
	//logger setup
	cfg, err := config.LoadConfig(filepath.Join("./../../integration_test", "conf", "test.yml"))
	if err != nil {
		fmt.Printf("error occured while loading config. error: %v", err)
	}
	logger.InitAppLogger(cfg.Logger)

	// test execution
	code := m.Run()

	os.Exit(code)
}
