package viper

import (
	"log"

	"github.com/spf13/viper"
)

func Setup() {
  viper.SetConfigName("config")
  viper.AddConfigPath(".")
  if err := viper.ReadInConfig(); err != nil {
    log.Fatalf("failed to read config: %v", err)
  }
}
