package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Host       string   `mapstructure:"host"`
  DBName string `mapstructure:"dbname"`
  User_DB string `mapstructure:"userdb"`
  PasswordDB string `mapstructure:"passworddb"`
	Admins     []int64 `mapstructure:"admins"`
	TgApiToken string   `mapstructure:"tg_api_token"`
  FSHost string `mapstructure:"fs_host"`
}

func InitConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("Error: config file is not found: %w", err)
		}
		return nil, fmt.Errorf("Error: init config: %w", err)
	}

	var cfg Config
	viper.ReadInConfig()
	err := viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatalf("Error: unable to decode into struct, %v", err)
	}
  fmt.Println(cfg)
	return &cfg, nil
}
