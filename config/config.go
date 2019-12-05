package config

import (
	"fmt"
	"runtime"

	"path/filepath"

	"github.com/spf13/viper"
)

var cfg *Config

type (
	Config struct {
		ReleaseMode             bool
		Postgres                PsqlConfig `json:"postgres"`
		Port                    int        `json:"port"`
		CertFile                string     `json:"certFile"`
		KeyFile                 string     `json:"keyFile"`
		RepeatCancellationEvery int        `json:"repeatCancellationEvery"`
		CancellationSelfRepeat  bool       `json:"cancellationSelfRepeat"`
	}

	PsqlConfig struct {
		Host     string `json:"host"`
		Username string `json:"username"`
		Password string `json:"password"`
		Database string `json:"database"`
		Port     int    `json:"port"`
	}
)

func GetConfig() Config {
	return *cfg
}

func GetPostgresConnection() string {
	return fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable",
		cfg.Postgres.Username,
		cfg.Postgres.Password,
		cfg.Postgres.Database,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
	)
}

func init() {
	var c Config
	viper.SetConfigName("config")
	viper.AddConfigPath(getConfigPath())
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf("Fatal error config file: %s \n", err))
	}
	err = viper.Unmarshal(&c)
	if err != nil {
		panic(fmt.Sprintf("Fatal error config file: %s \n", err))
	}
	cfg = &c
}

func getConfigPath() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filepath.Dir(filename))
}
