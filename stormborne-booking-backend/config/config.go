package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Server       ServerConfig
	Database     DbConfig
	MovieGateway MovieGatewayConfig
	Logger       LoggerConfig
	Token        TokenConfig

	DB_PASSWORD string
}

type ServerConfig struct {
	Port         int `yaml:"port"`
	ReadTimeout  int
	WriteTimeout int
	GineMode     string
	WebClientResetPasswordBaseUrl string
}

type DbConfig struct {
	Host          string `yaml:"host"`
	Port          int    `yaml:"port"`
	Name          string `yaml:"name"`
	User          string `yaml:"user"`
	Password      string `yaml:"password"`
	MigrationPath string
}

type MovieGatewayConfig struct {
	MovieServiceHost string `yaml:"movieServiceHost"`
}

type LoggerConfig struct {
	Level string
}

type TokenConfig struct {
	TTL    int    `yaml:"ttl"`
	SECRET string `yaml:"secret"`
}

func LoadConfig(configFile string) (AppConfig, error) {
	viper.SetConfigFile(configFile)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return AppConfig{}, fmt.Errorf("failed to load the config: %w", err)
	}

	for _, k := range viper.AllKeys() {
		v := viper.GetString(k)
		viper.Set(k, os.ExpandEnv(v))
	}

	var c AppConfig
	if err := viper.Unmarshal(&c); err != nil {
		return AppConfig{}, fmt.Errorf("failed to parse the config. %w", err)
	}
	return c, nil
}