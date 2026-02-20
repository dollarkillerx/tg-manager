package client

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tg-manager/pkg/common/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func PostgresClient(conf config.PostgresConfiguration) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		conf.Host, conf.Port, conf.User, conf.Password, conf.DBName, conf.SSLMode)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
	})
}
