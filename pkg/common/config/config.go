package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type ServiceConfiguration struct {
	Port  string `mapstructure:"Port"`
	Debug bool   `mapstructure:"Debug"`
}

type PostgresConfiguration struct {
	Host     string `mapstructure:"Host"`
	Port     int    `mapstructure:"Port"`
	User     string `mapstructure:"User"`
	Password string `mapstructure:"Password"`
	DBName   string `mapstructure:"DBName"`
	SSLMode  string `mapstructure:"SSLMode"`
}

type LoggerConfig struct {
	Filename string `mapstructure:"Filename"`
	MaxSize  int    `mapstructure:"MaxSize"` // MB
}

type TelegramConfiguration struct {
	AppID   int    `mapstructure:"AppID"`
	AppHash string `mapstructure:"AppHash"`
}

func InitConfiguration(configName string, configPaths []string, config interface{}) error {
	vp := viper.New()
	vp.SetConfigName(configName)
	vp.AutomaticEnv()

	for _, p := range configPaths {
		vp.AddConfigPath(p)
	}

	if err := vp.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return errors.WithStack(err)
		}
	}

	if err := vp.Unmarshal(config); err != nil {
		return errors.WithStack(err)
	}

	for _, key := range vp.AllKeys() {
		if err := vp.BindEnv(key); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
