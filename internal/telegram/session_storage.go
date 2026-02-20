package telegram

import (
	"context"
	"errors"

	"github.com/gotd/td/session"
	"github.com/tg-manager/internal/storage"
	"gorm.io/gorm"
)

type PostgresSessionStorage struct {
	db *gorm.DB
}

func NewPostgresSessionStorage(db *gorm.DB) *PostgresSessionStorage {
	return &PostgresSessionStorage{db: db}
}

func (s *PostgresSessionStorage) LoadSession(ctx context.Context) ([]byte, error) {
	var sess storage.TelegramSession
	if err := s.db.WithContext(ctx).First(&sess).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, session.ErrNotFound
		}
		return nil, err
	}
	return sess.Data, nil
}

func (s *PostgresSessionStorage) StoreSession(ctx context.Context, data []byte) error {
	var sess storage.TelegramSession
	result := s.db.WithContext(ctx).First(&sess)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}
	sess.Data = data
	return s.db.WithContext(ctx).Save(&sess).Error
}
