package conf

import "github.com/tg-manager/pkg/common/config"

type Config struct {
	ServiceConfiguration  config.ServiceConfiguration  `mapstructure:"ServiceConfiguration"`
	SQLiteConfiguration   config.SQLiteConfiguration   `mapstructure:"SQLiteConfiguration"`
	LoggerConfiguration   config.LoggerConfig          `mapstructure:"LoggerConfiguration"`
	TelegramConfiguration config.TelegramConfiguration `mapstructure:"TelegramConfiguration"`
}
