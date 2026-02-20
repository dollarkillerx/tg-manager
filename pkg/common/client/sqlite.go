package client

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/tg-manager/pkg/common/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func SQLiteClient(conf config.SQLiteConfiguration) (*gorm.DB, error) {
	_ = os.MkdirAll(filepath.Dir(conf.Path), 0o755)

	dsn := conf.Path + "?_journal_mode=WAL"
	return gorm.Open(sqlite.Open(dsn), &gorm.Config{
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
