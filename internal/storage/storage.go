// Package storage реализует интерфейс хранения данных
package storage

import (
	"errors"
	"github.com/gerasimovpavel/shortener.git/internal/config"
)

// ErrDataConflict Ошибка конфликта дубликата данных
var ErrDataConflict = errors.New("дубликат данных")

// Storage Инткрфейс хранилища
type Storage interface {
	Get(shortURL string) (*URLData, error)
	Post(data *URLData) error
	PostBatch(urls []*URLData) error
	FindByOriginalURL(originalURL string) (*URLData, error)
	Ping() error
	Close() error
	GetUserURL(userID string) ([]*URLData, error)
	DeleteUserURL(urls []*URLData) error
}

// Stor Глобальная переменная для работы с хранилищем ссылок
var Stor Storage

// URLData Структура хранящая информацию о ссылке
type URLData struct {
	UUID        string `json:"uuid,omitempty" db:"uuid"`
	CorrID      string `json:"correlation_id,omitempty"`
	ShortURL    string `json:"short_url,omitempty" db:"shortURL"`
	OriginalURL string `json:"original_url,omitempty" db:"originalURL"`
	UserID      string `json:"user_id,omitempty" db:"userID"`
	DeletedFlag bool   `json:"-" db:"is_deleted"`
}

// NewStorage создание нового хранилища
func NewStorage() (Storage, error) {
	if config.Options.DatabaseDSN != "" {
		return NewPostgreWorker(config.Options.DatabaseDSN)
	}
	if config.Options.FileStoragePath != "" {
		return NewFileWorker(config.Options.FileStoragePath)
	}
	return NewMemWorker()
}
