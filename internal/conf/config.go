package conf

import "github.com/tg-manager/pkg/common/config"

type Config struct {
	ServiceConfiguration  config.ServiceConfiguration  `mapstructure:"ServiceConfiguration"`
	PostgresConfiguration config.PostgresConfiguration `mapstructure:"PostgresConfiguration"`
	LoggerConfiguration   config.LoggerConfig          `mapstructure:"LoggerConfiguration"`
	TelegramConfiguration config.TelegramConfiguration `mapstructure:"TelegramConfiguration"`
}
