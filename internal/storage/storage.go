package storage

import "gorm.io/gorm"

type Storage struct {
	db *gorm.DB
}

func NewStorage(db *gorm.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) AutoMigrate() error {
	return s.db.AutoMigrate(&ForwardRule{}, &ForwardLog{}, &TelegramSession{})
}

func (s *Storage) GetDB() *gorm.DB { return s.db }
